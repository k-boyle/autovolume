package plex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"text/template"
)

var Url = mustParse("https://clients.plex.tv")

type ApiClient struct {
	plexTv *http.Client
	pms    *http.Client

	GetUser               Endpoint0[User]
	ListDevices           Endpoint0[[]Device]
	GetIdentity           Endpoint0[Identity]
	GetAllLibrarySections Endpoint0[LibrarySections]
	GetLibrarySection     Endpoint[GetLibrarySectionParams, LibrarySection]
	GetLibraryChildren    Endpoint[GetLibraryChildrenParams, LibraryChildren]
	GetLibraryMetadata    Endpoint[GetLibraryMetadataParams, LibraryMetadatas]
}

func NewApiClient(plexTv *http.Client, pms *http.Client) *ApiClient {
	c := &ApiClient{
		plexTv: plexTv,
		pms:    pms,
	}
	c.GetUser = noParams(c.plexGet[struct{}, User]("/api/v2/user"))
	c.ListDevices = noParams(c.plexGet[struct{}, []Device]("/api/v2/devices"))
	c.GetIdentity = noParams(c.pmsGet[struct{}, Identity]("/identity"))
	c.GetAllLibrarySections = noParams(c.pmsGet[struct{}, LibrarySections]("/library/sections/all"))
	c.GetLibrarySection = c.pmsGet[GetLibrarySectionParams, LibrarySection]("/library/sections/{{ path .Id }}/all")
	c.GetLibraryChildren = c.pmsGet[GetLibraryChildrenParams, LibraryChildren]("/library/metadata/{{ path .RatingKey }}/children")
	c.GetLibraryMetadata = c.pmsGet[GetLibraryMetadataParams, LibraryMetadatas]("/library/metadata/{{ path .RatingKey }}")
	return c
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

type Endpoint[PARAMS any, RESPONSE any] func(context.Context, PARAMS, ...Override) (RESPONSE, error)
type Endpoint0[RESPONSE any] func(context.Context, ...Override) (RESPONSE, error)

func (c *ApiClient) endpoint[PARAMS any, RESPONSE any](client *http.Client, method, pattern string) Endpoint[PARAMS, RESPONSE] {
	route := createRoute[PARAMS](pattern)
	return func(ctx context.Context, params PARAMS, overrides ...Override) (RESPONSE, error) {
		path, query, err := route.build(params)
		if err != nil {
			var zero RESPONSE
			return zero, err
		}
		return c.do[RESPONSE](ctx, client, method, path, query, overrides)
	}
}

func (c *ApiClient) plexGet[PARAMS any, RESPONSE any](pattern string) Endpoint[PARAMS, RESPONSE] {
	return c.endpoint[PARAMS, RESPONSE](c.plexTv, http.MethodGet, pattern)
}

func (c *ApiClient) pmsGet[PARAMS any, RESPONSE any](pattern string) Endpoint[PARAMS, RESPONSE] {
	return unwrap(c.endpoint[PARAMS, mediaContainer[RESPONSE]](c.pms, http.MethodGet, pattern))
}

func unwrap[PARAMS any, RESPONSE any](e Endpoint[PARAMS, mediaContainer[RESPONSE]]) Endpoint[PARAMS, RESPONSE] {
	return func(ctx context.Context, params PARAMS, overrides ...Override) (RESPONSE, error) {
		r, err := e(ctx, params, overrides...)
		return r.Value, err
	}
}

func noParams[RESPONSE any](e Endpoint[struct{}, RESPONSE]) Endpoint0[RESPONSE] {
	return func(ctx context.Context, overrides ...Override) (RESPONSE, error) {
		return e(ctx, struct{}{}, overrides...)
	}
}

type queryParameter struct {
	key       string
	omitEmpty bool
	index     []int
	format    func(reflect.Value) string
}

type route[PARAMS any] struct {
	path            *template.Template
	queryParameters []queryParameter
}

func createRoute[PARAMS any](pattern string) route[PARAMS] {
	tmpl := template.Must(template.New(pattern).Funcs(template.FuncMap{"path": url.PathEscape}).Parse(pattern))
	if err := tmpl.Execute(io.Discard, *new(PARAMS)); err != nil {
		panic(err)
	}

	var queryParameters []queryParameter
	for field := range reflect.TypeFor[PARAMS]().Fields() {
		tag, ok := field.Tag.Lookup("query")
		if !ok {
			continue
		}

		key, options, _ := strings.Cut(tag, ",")
		if key == "" {
			panic(fmt.Errorf("%s: empty query tag on %s", pattern, field.Name))
		}

		queryParameters = append(queryParameters, queryParameter{
			key:       key,
			omitEmpty: options == "omitempty",
			index:     field.Index,
			format:    formatter(field.Type),
		})
	}

	return route[PARAMS]{
		path:            tmpl,
		queryParameters: queryParameters,
	}
}

func formatter(t reflect.Type) func(reflect.Value) string {
	switch t.Kind() {
	case reflect.String:
		return reflect.Value.String
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return func(v reflect.Value) string {
			return strconv.FormatInt(v.Int(), 10)
		}
	case reflect.Bool:
		return func(v reflect.Value) string {
			if v.Bool() {
				return "1"
			}
			return "0"
		}
	default:
		panic(fmt.Errorf("unsupported query parameter type: %s", t))
	}
}

func (r route[PARAMS]) build(params PARAMS) (string, url.Values, error) {
	path := strings.Builder{}
	if err := r.path.Execute(&path, params); err != nil {
		return "", nil, err
	}

	query := url.Values{}
	v := reflect.ValueOf(params)
	for _, parameter := range r.queryParameters {
		field := v.FieldByIndex(parameter.index)
		if parameter.omitEmpty && field.IsZero() {
			continue
		}

		query.Set(parameter.key, parameter.format(field))
	}

	return path.String(), query, nil
}

func (c *ApiClient) do[RESPONSE any](ctx context.Context, client *http.Client, method, path string, query url.Values, overrides []Override) (RESPONSE, error) {
	var zero RESPONSE

	full, err := url.Parse(path)
	if err != nil {
		return zero, err
	}

	var token string
	for _, override := range overrides {
		if override.Url != "" {
			base, err := url.Parse(override.Url)
			if err != nil {
				return zero, err
			}
			full = base.JoinPath(path)
		}
		if override.Token != "" {
			token = override.Token
		}
	}
	full.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, method, full.String(), nil)
	if err != nil {
		return zero, err
	}
	if token != "" {
		req.Header.Set("X-Plex-Token", token)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-Plex-Product", "AutoVolume")

	rsp, err := client.Do(req)
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

	var r RESPONSE
	if err := json.Unmarshal(body, &r); err != nil {
		return zero, err
	}

	return r, nil
}

func (c *ApiClient) GetThumb(ctx context.Context, thumb, width, height string) (*http.Response, error) {
	u := &url.URL{Path: "/photo/:/transcode"}
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
	return c.pms.Do(req)
}
