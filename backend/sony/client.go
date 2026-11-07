package sony

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

type Client struct {
	http *http.Client

	GetVolumeInformation    Endpoint[GetVolumeInformationParams, []GetVolumeInformationResponse]
	GetInterfaceInformation Endpoint0[GetInterfaceInformationResponse]
	GetPowerStatus          Endpoint0[GetPowerStatusResponse]
	SetAudioVolume          Endpoint[SetVolumeInformationParams, SetVolumeInformationResponse]
}

type Endpoint[PARAMS any, RESPONSE any] func(context.Context, PARAMS, ...Override) (RESPONSE, error)
type Endpoint0[RESPONSE any] func(context.Context, ...Override) (RESPONSE, error)

type Override struct {
	Url string
	Psk *string
}

type Error struct {
	Code    int
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("sony error %d: %s", e.Code, e.Message)
}

func (e *Error) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &[]any{&e.Code, &e.Message})
}

func NewClient(http *http.Client) *Client {
	c := &Client{
		http: http,
	}
	c.GetVolumeInformation = c.endpoint[GetVolumeInformationParams, []GetVolumeInformationResponse]("audio", "getVolumeInformation", "1.1")
	c.GetInterfaceInformation = c.endpoint0[GetInterfaceInformationResponse]("system", "getInterfaceInformation", "1.0")
	c.GetPowerStatus = c.endpoint0[GetPowerStatusResponse]("system", "getPowerStatus", "1.1")
	c.SetAudioVolume = c.endpoint[SetVolumeInformationParams, SetVolumeInformationResponse]("audio", "setAudioVolume", "1.0")
	return c
}

func (c *Client) endpoint[PARAMS any, RESPONSE any](service, method, version string) Endpoint[PARAMS, RESPONSE] {
	return func(ctx context.Context, params PARAMS, overrides ...Override) (RESPONSE, error) {
		return c.call[RESPONSE](ctx, service, method, version, []any{params}, overrides)
	}
}

func (c *Client) endpoint0[RESPONSE any](service, method, version string) Endpoint0[RESPONSE] {
	return func(ctx context.Context, overrides ...Override) (RESPONSE, error) {
		return c.call[RESPONSE](ctx, service, method, version, []any{}, overrides)
	}
}

type request struct {
	Method  string `json:"method"`
	Id      int    `json:"id"`
	Version string `json:"version"`
	Params  []any  `json:"params"`
}

type response struct {
	Result []json.RawMessage `json:"result"`
	Error  *Error            `json:"error"`
}

func (c *Client) call[RESPONSE any](ctx context.Context, service, method, version string, params []any, overrides []Override) (RESPONSE, error) {
	var zero RESPONSE

	endpoint := &url.URL{Path: "/sony/" + service}
	var psk string
	for _, override := range overrides {
		if override.Url != "" {
			base, err := url.Parse(override.Url)
			if err != nil {
				return zero, err
			}
			endpoint = base.JoinPath("/sony", service)
		}
		if override.Psk != nil {
			psk = *override.Psk
		}
	}

	body, err := json.Marshal(request{Method: method, Id: 1, Version: version, Params: params})
	if err != nil {
		return zero, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return zero, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if psk != "" {
		req.Header.Set("X-Auth-PSK", psk)
	}

	rsp, err := c.http.Do(req)
	if err != nil {
		return zero, err
	}
	defer func(body io.ReadCloser) {
		if err := body.Close(); err != nil {
			slog.Error("failed to close response body", slog.Any("error", err))
		}
	}(rsp.Body)

	var r response
	if err := json.NewDecoder(rsp.Body).Decode(&r); err != nil {
		return zero, err
	}
	if r.Error != nil {
		return zero, r.Error
	}
	if rsp.StatusCode != http.StatusOK {
		return zero, fmt.Errorf("invalid status code: %d", rsp.StatusCode)
	}
	if len(r.Result) == 0 {
		return zero, nil
	}

	var result RESPONSE
	if err := json.Unmarshal(r.Result[0], &result); err != nil {
		return zero, err
	}
	return result, nil
}
