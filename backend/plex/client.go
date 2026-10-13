package plex

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var clientsUrl *url.URL
var Url *url.URL

func init() {
	clientsUrl, _ = url.Parse("https://clients.plex.tv")
	Url, _ = url.Parse("https://plex.tv")
}

type Client struct {
	client   *http.Client
	pub      ed25519.PublicKey
	prv      ed25519.PrivateKey
	clientId string

	tokenAccessor func(context.Context) (string, error)
}

func New(client *http.Client, clientId string, pub ed25519.PublicKey, prv ed25519.PrivateKey, tokenAccessor func(context.Context) (string, error)) (*Client, error) {
	return &Client{
		client:        client,
		pub:           pub,
		prv:           prv,
		clientId:      clientId,
		tokenAccessor: tokenAccessor,
	}, nil
}

func (c *Client) CreatePin(ctx context.Context) (*Pin, error) {
	b, err := json.Marshal(createPinBody{
		Jwk: jwk{
			Kty: "OKP",
			Crv: "Ed25519",
			X:   base64.RawURLEncoding.EncodeToString(c.pub),
			Kid: "autovolume",
			Alg: "EdDSA",
		},
		Strong: true,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", clientsUrl.JoinPath("api", "v2", "pins").String(), bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	return c.executeRequest[Pin](req)
}

func (c *Client) ExchangePin(ctx context.Context, pinId int) (*Pin, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.MapClaims{
		"aud": "plex.tv",
		"iss": c.clientId,
		"iat": now.Unix(),
		"exp": now.Add(time.Minute * 5).Unix(),
	})
	token.Header["kid"] = "autovolume"
	deviceJwt, err := token.SignedString(c.prv)
	if err != nil {
		return nil, err
	}

	uri := clientsUrl.JoinPath("api", "v2", "pins", strconv.Itoa(pinId))
	query := uri.Query()
	query.Set("deviceJWT", deviceJwt)
	uri.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, "GET", uri.String(), nil)
	if err != nil {
		return nil, err
	}

	return c.executeRequest[Pin](req)
}

func (c *Client) GetNonce(ctx context.Context) (*Nonce, error) {
	uri := clientsUrl.JoinPath("api", "v2", "auth", "nonce")
	req, err := http.NewRequestWithContext(ctx, "GET", uri.String(), nil)
	if err != nil {
		return nil, err
	}

	return c.executeRequest[Nonce](req)
}

func (c *Client) ExchangePlexToken(ctx context.Context, nonce *Nonce) (*Pin, error) {
	uri := clientsUrl.JoinPath("api", "v2", "auth", "token")
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.MapClaims{
		"nonce": nonce.Value,
		"iss":   c.clientId,
		"iat":   now.Unix(),
		"exp":   now.Add(time.Minute * 5).Unix(),
		"aud":   "plex.tv",
		"scope": "username,email,friendly_name",
	})
	token.Header["kid"] = "autovolume"
	token.Header["alg"] = "EdDSA"
	token.Header["typ"] = "JWT"

	tokenString, err := token.SignedString(c.prv)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(exchangeAuthBody{Jwt: tokenString})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", uri.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	return c.executeRequest[Pin](req)
}

func (c *Client) GetUsers(ctx context.Context) (*Users, error) {
	uri := Url.JoinPath("api", "v2", "home", "users")
	req, err := http.NewRequestWithContext(ctx, "GET", uri.String(), nil)
	if err != nil {
		return nil, err
	}

	return c.executeAuthenticatedRequest[Users](req)
}

func (c *Client) executeAuthenticatedRequest[T any](req *http.Request) (*T, error) {
	token, err := c.tokenAccessor(req.Context())
	if err != nil {
		return nil, err
	}
	if len(token) == 0 {
		return nil, fmt.Errorf("no token found")
	}
	req.Header.Set("X-Plex-Token", token)
	return c.executeRequest[T](req)
}

func (c *Client) executeRequest[T any](req *http.Request) (*T, error) {
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-Plex-Client-Identifier", c.clientId)
	req.Header.Add("X-Plex-Product", "AutoVolume")

	rsp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			log.Printf("failed to close response body: %v", err)
		}
	}(rsp.Body)

	if rsp.StatusCode < http.StatusOK || rsp.StatusCode >= 300 {
		return nil, fmt.Errorf("invalid status code: %d", rsp.StatusCode)
	}

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}

	t := new(T)
	if err := json.Unmarshal(body, t); err != nil {
		return nil, err
	}

	return t, nil
}
