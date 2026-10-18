package plex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

var AppUrl = mustParse("https://app.plex.tv")

type ApiClient struct {
	client        *http.Client
	clientId      string
	tokenAccessor func(context.Context) (string, error)

	GetUsers    Endpoint[Users]
	ListDevices Endpoint[[]Device]
	GetIdentity Endpoint[Identity]
}

func NewApiClient(client *http.Client, clientId string, tokenAccessor func(context.Context) (string, error)) *ApiClient {
	c := &ApiClient{
		client:        client,
		clientId:      clientId,
		tokenAccessor: tokenAccessor,
	}
	c.GetUsers = &endpoint[Users]{
		c:      c,
		path:   "/api/v2/home/users",
		method: "GET",
	}
	c.ListDevices = &endpoint[[]Device]{
		c:      c,
		path:   "/api/v2/devices",
		method: "GET",
	}
	c.GetIdentity = &mediaContainerEndpoint[Identity]{
		c:      c,
		path:   "/api/v2/identity",
		method: "GET",
	}
	return c
}

type Endpoint[R any] interface {
	Execute(context.Context, ...Override) (R, error)
}

type endpoint[R any] struct {
	c      *ApiClient
	path   string
	method string

	zero R
}

type mediaContainerEndpoint[R any] struct {
	endpoint[mediaContainer[R]]
}

type Override struct {
	Url string
}

func (e *mediaContainerEndpoint[R]) Execute(ctx context.Context, overrides ...Override) (R, error) {
	r, err := e.endpoint.Execute(ctx, overrides...)
	if err != nil {
		var zero R
		return zero, err
	}

	return r.Value, nil
}

func (e *endpoint[R]) Execute(ctx context.Context, overrides ...Override) (R, error) {
	token, err := e.c.tokenAccessor(ctx)
	if err != nil {
		return e.zero, err
	}

	u := AppUrl.JoinPath(e.path)
	for _, override := range overrides {
		u, err = u.Parse(override.Url)
		if err != nil {
			return e.zero, err
		}
		u = u.JoinPath(e.path)
	}

	req, err := http.NewRequestWithContext(ctx, e.method, u.String(), nil)
	if err != nil {
		return e.zero, err
	}
	req.Header.Set("X-Plex-Token", token)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-Plex-Client-Identifier", e.c.clientId)
	req.Header.Add("X-Plex-Product", "AutoVolume")

	rsp, err := e.c.client.Do(req)
	if err != nil {
		return e.zero, err
	}
	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			slog.Error("failed to close response body", slog.Any("error", err))
		}
	}(rsp.Body)

	if rsp.StatusCode < http.StatusOK || rsp.StatusCode >= 300 {
		return e.zero, fmt.Errorf("invalid status code: %d", rsp.StatusCode)
	}

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return e.zero, err
	}

	var r R
	if err := json.Unmarshal(body, &r); err != nil {
		return e.zero, err
	}

	return r, nil
}
