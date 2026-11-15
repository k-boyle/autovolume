package plex

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestExchangePinSendsSignedDeviceJwt(t *testing.T) {
	credentials, err := NewCredentials()
	if err != nil {
		t.Fatal(err)
	}

	var deviceJwt string
	client := NewAuthClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		deviceJwt = req.URL.Query().Get("deviceJWT")
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"id": 1, "authToken": "token"}`)),
		}, nil
	})})

	pin, err := client.ExchangePin(t.Context(), credentials, 1)
	if err != nil {
		t.Fatal(err)
	}
	if pin.AuthToken == nil || *pin.AuthToken != "token" {
		t.Errorf("auth token = %v, want %q", pin.AuthToken, "token")
	}

	token, err := jwt.Parse(deviceJwt, func(*jwt.Token) (any, error) {
		return credentials.Key.Public(), nil
	},
		jwt.WithValidMethods([]string{"EdDSA"}),
		jwt.WithAudience("plex.tv"),
		jwt.WithIssuer(credentials.ClientId),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if token.Header["kid"] != credentials.KeyId {
		t.Errorf("kid = %v, want %q", token.Header["kid"], credentials.KeyId)
	}
}

func TestTokenExpiry(t *testing.T) {
	exp := time.Now().Add(time.Hour).Truncate(time.Second)
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"exp": exp.Unix()}).SignedString([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}

	got, err := TokenExpiry(token)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(exp) {
		t.Errorf("expiry = %v, want %v", got, exp)
	}
}
