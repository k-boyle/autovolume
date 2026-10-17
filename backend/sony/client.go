package sony

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type Client struct {
	http        *http.Client
	urlAccessor func(ctx context.Context) (string, string, error)

	GetVolumeInformation    Endpoint[GetVolumeInformationParams, GetVolumeInformationResponse]
	GetInterfaceInformation Endpoint[GetInterfaceInformationParams, GetInterfaceInformationResponse]
	GetPowerStatus          Endpoint[GetPowerStatusParams, GetPowerStatusResponse]
}

func NewClient(http *http.Client, urlAccessor func(ctx context.Context) (string, string, error)) *Client {
	c := &Client{
		http:        http,
		urlAccessor: urlAccessor,
	}
	c.GetVolumeInformation = Endpoint[GetVolumeInformationParams, GetVolumeInformationResponse]{
		c:       c,
		path:    "/sony/audio",
		method:  "getVolumeInformation",
		id:      33,
		version: "1.1",
	}
	c.GetInterfaceInformation = Endpoint[GetInterfaceInformationParams, GetInterfaceInformationResponse]{
		c:       c,
		path:    "/sony/system",
		method:  "getInterfaceInformation",
		id:      33,
		version: "1.0",
	}
	c.GetPowerStatus = Endpoint[GetPowerStatusParams, GetPowerStatusResponse]{
		c:       c,
		path:    "/sony/system",
		method:  "getPowerStatus",
		id:      50,
		version: "1.1",
	}
	return c
}

type Endpoint[R1 any, R2 any] struct {
	c       *Client
	path    string
	method  string
	id      int
	version string
}

type Override struct {
	Url string
	Psk *string
}

func (e Endpoint[R1, R2]) Execute(ctx context.Context, params *R1, overrides ...Override) (*R2, error) {
	reqBody, err := json.Marshal(requestBody[R1]{
		Method:  e.method,
		Id:      e.id,
		Version: e.version,
		Params:  []*R1{params},
	})
	if err != nil {
		return nil, err
	}

	url, psk, err := e.c.urlAccessor(ctx)
	if err != nil {
		return nil, err
	}
	for _, override := range overrides {
		url = override.Url
		if override.Psk != nil {
			psk = *override.Psk
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url+e.path, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	if len(psk) > 0 {
		req.Header.Add("PSK", psk)
	}

	rsp, err := e.c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			slog.Error("failed to close response body", slog.Any("error", err))
		}
	}(rsp.Body)

	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid status code: %d", rsp.StatusCode)
	}

	rspBody := new(responseBody[R2])
	if err := json.NewDecoder(rsp.Body).Decode(rspBody); err != nil {
		return nil, err
	}

	return rspBody.Result[0], nil
}

type requestBody[P any] struct {
	Method  string `json:"method"`
	Id      int    `json:"id"`
	Version string `json:"version"`
	Params  []*P   `json:"params"`
}

type responseBody[P any] struct {
	Result []*P `json:"result"`
	Id     int  `json:"id"`
}
