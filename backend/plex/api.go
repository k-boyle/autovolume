package plex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

var AppUrl = mustParse("https://app.plex.tv")

type ApiClient struct {
	client       *http.Client
	clientId     string
	plexAccessor Accessor
	pmsAccessor  Accessor

	GetUsers    Endpoint[Users]
	ListDevices Endpoint[[]Device]
	GetIdentity Endpoint[Identity]
}

type Accessor func(context.Context) (*url.URL, string, error)

func NewApiClient(client *http.Client, clientId string, plexAccessor Accessor, pmsAccessor Accessor) *ApiClient {
	c := &ApiClient{
		client:       client,
		clientId:     clientId,
		plexAccessor: plexAccessor,
		pmsAccessor:  pmsAccessor,
	}
	c.GetUsers = &plexEndpoint[Users]{
		c:      c,
		path:   "/api/v2/home/users",
		method: "GET",
	}
	c.ListDevices = &plexEndpoint[[]Device]{
		c:      c,
		path:   "/api/v2/devices",
		method: "GET",
	}
	c.GetIdentity = &pmsEndpoint[Identity]{
		c:      c,
		path:   "/identity",
		method: "GET",
	}
	return c
}

type Endpoint[R any] interface {
	Execute(context.Context, ...Override) (R, error)
}

type plexEndpoint[R any] struct {
	c      *ApiClient
	path   string
	method string

	zero R
}

type pmsEndpoint[R any] struct {
	plexEndpoint[mediaContainer[R]]
}

type Override struct {
	Url   string
	Token string
}

func (e *pmsEndpoint[R]) Execute(ctx context.Context, overrides ...Override) (R, error) {
	r, err := e.plexEndpoint.execute(ctx, e.c.pmsAccessor, overrides...)
	if err != nil {
		var zero R
		return zero, err
	}

	return r.Value, nil
}

func (e *plexEndpoint[R]) Execute(ctx context.Context, overrides ...Override) (R, error) {
	return e.execute(ctx, e.c.plexAccessor, overrides...)
}

func (e *plexEndpoint[R]) execute(ctx context.Context, accessor Accessor, overrides ...Override) (R, error) {
	u, token, err := accessor(ctx)
	if err != nil {
		return e.zero, err
	}

	u = u.JoinPath(e.path)
	for _, override := range overrides {
		if override.Url != "" {
			u, err = u.Parse(override.Url)
			if err != nil {
				return e.zero, err
			}
			u = u.JoinPath(e.path)
		}
		if override.Token != "" {
			token = override.Token
		}
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
