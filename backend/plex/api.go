package plex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"text/template"
)

var AppUrl = mustParse("https://app.plex.tv")

type ApiClient struct {
	client       *http.Client
	clientId     string
	plexAccessor Accessor
	pmsAccessor  Accessor

	GetUsers              Endpoint[Users]
	ListDevices           Endpoint[[]Device]
	GetIdentity           Endpoint[Identity]
	GetAllLibrarySections Endpoint[LibrarySections]
	GetLibrarySection     EndpointWithPath[LibrarySection, GetLibrarySectionPathParams]
	GetLibraryChildren    EndpointWithPath[LibraryChildren, GetLibraryChildrenParams]
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
	c.GetAllLibrarySections = &pmsEndpoint[LibrarySections]{
		c:      c,
		path:   "/library/sections/all",
		method: "GET",
	}
	c.GetLibrarySection = &pmsPathEndpoint[LibrarySection, GetLibrarySectionPathParams]{
		c:      c,
		path:   mustTemplate("get_library_section", "/library/sections/{{ .Id }}/all"),
		method: "GET",
	}
	c.GetLibraryChildren = &pmsPathEndpoint[LibraryChildren, GetLibraryChildrenParams]{
		c:      c,
		path:   mustTemplate("get_library_children", "/library/metadata/{{ .RatingKey }}/children"),
		method: "GET",
	}
	return c
}

type Endpoint[R any] interface {
	Execute(context.Context, ...Override) (R, error)
}

type EndpointWithPath[R any, P any] interface {
	Execute(context.Context, P, ...Override) (R, error)
}

type plexEndpoint[R any] struct {
	c      *ApiClient
	path   string
	method string

	zero R
}

type pmsEndpoint[R any] struct {
	c      *ApiClient
	path   string
	method string

	zero R
}

type pmsPathEndpoint[R any, P any] struct {
	c      *ApiClient
	path   *template.Template
	method string

	zero R
}

type Override struct {
	Url   string
	Token string
}

type StatusError struct {
	Code int
}

func (e StatusError) Error() string {
	return fmt.Sprintf("invalid status code: %s", http.StatusText(e.Code))
}

func (e *pmsEndpoint[R]) Execute(ctx context.Context, overrides ...Override) (R, error) {
	r, err := e.c.execute[mediaContainer[R]](ctx, e.path, e.method, e.c.pmsAccessor, overrides)
	if err != nil {
		var zero R
		return zero, err
	}

	return r.Value, nil
}

func (e *plexEndpoint[R]) Execute(ctx context.Context, overrides ...Override) (R, error) {
	return e.c.execute[R](ctx, e.path, e.method, e.c.plexAccessor, overrides)
}

func (p pmsPathEndpoint[R, P]) Execute(ctx context.Context, pathParams P, overrides ...Override) (R, error) {
	path := strings.Builder{}
	if err := p.path.Execute(&path, pathParams); err != nil {
		var zero R
		return zero, err
	}

	r, err := p.c.execute[mediaContainer[R]](ctx, path.String(), p.method, p.c.pmsAccessor, overrides)
	if err != nil {
		var zero R
		return zero, err
	}

	return r.Value, nil
}

func (c *ApiClient) execute[R any](ctx context.Context, path, method string, accessor Accessor, overrides []Override) (R, error) {
	var zero R
	base, token, err := accessor(ctx)
	if err != nil {
		return zero, err
	}

	base = base.JoinPath(path)
	for _, override := range overrides {
		if override.Url != "" {
			base, err = base.Parse(override.Url)
			if err != nil {
				return zero, err
			}
			base = base.JoinPath(path)
		}
		if override.Token != "" {
			token = override.Token
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, base.String(), nil)
	if err != nil {
		return zero, err
	}
	req.Header.Set("X-Plex-Token", token)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-Plex-Client-Identifier", c.clientId)
	req.Header.Add("X-Plex-Product", "AutoVolume")

	rsp, err := c.client.Do(req)
	if err != nil {
		return zero, err
	}
	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			slog.Error("failed to close response body", slog.Any("error", err))
		}
	}(rsp.Body)

	if rsp.StatusCode < http.StatusOK || rsp.StatusCode >= 300 {
		return zero, &StatusError{Code: rsp.StatusCode}
	}

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return zero, err
	}

	var r R
	if err := json.Unmarshal(body, &r); err != nil {
		return zero, err
	}

	return r, nil
}

func (c *ApiClient) GetThumb(ctx context.Context, thumb, width, height string) (*http.Response, error) {
	base, token, err := c.pmsAccessor(ctx)
	if err != nil {
		return nil, err
	}
	u := base.JoinPath("/photo/:/transcode")
	u.RawQuery = url.Values{
		"url":     {thumb},
		"width":   {width},
		"height":  {height},
		"minSize": {"1"},
		"upscale": {"1"},
	}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Plex-Token", token)
	return c.client.Do(req)
}

func mustTemplate(name, s string) *template.Template {
	tmpl, err := template.New(name).Parse(s)
	if err != nil {
		panic(err)
	}
	return tmpl
}
