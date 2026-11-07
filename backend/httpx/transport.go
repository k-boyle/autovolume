package httpx

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type Target struct {
	Base   *url.URL
	Header http.Header
}

type Resolver func(ctx context.Context) (*Target, error)

type Transport struct {
	Resolve Resolver
	Next    http.RoundTripper
}

func NewClient(resolve Resolver) *http.Client {
	return &http.Client{
		Transport: &Transport{
			Resolve: resolve,
			Next:    http.DefaultTransport,
		},
	}
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	target, err := t.Resolve(req.Context())
	if err != nil {
		closeBody(req)
		return nil, err
	}

	req = req.Clone(req.Context())
	if req.URL.Host == "" {
		if target.Base == nil {
			closeBody(req)
			return nil, fmt.Errorf("no base url configured for %s", req.URL.Path)
		}
		base := *target.Base
		if base.Path == "" {
			base.Path = "/"
		}
		u := base.JoinPath(req.URL.EscapedPath())
		u.RawQuery = req.URL.RawQuery
		req.URL = u
	}

	for key, values := range target.Header {
		if req.Header.Get(key) != "" {
			continue
		}
		for _, value := range values {
			if value != "" {
				req.Header.Add(key, value)
			}
		}
	}

	return t.Next.RoundTrip(req)
}

func closeBody(req *http.Request) {
	if req.Body != nil {
		_ = req.Body.Close()
	}
}
