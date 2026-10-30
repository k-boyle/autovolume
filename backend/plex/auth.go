package plex

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthClient struct {
	client *http.Client
}

type Credentials struct {
	ClientId string
	KeyId    string
	Key      ed25519.PrivateKey
}

func NewAuthClient(client *http.Client) *AuthClient {
	return &AuthClient{
		client: client,
	}
}

func NewCredentials() (*Credentials, error) {
	_, prv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	return &Credentials{
		ClientId: rand.Text(),
		KeyId:    thumbprint(prv.Public().(ed25519.PublicKey)),
		Key:      prv,
	}, nil
}

func CredentialsFromSeed(clientId, keyId string, seed []byte) (*Credentials, error) {
	if len(seed) != ed25519.SeedSize || keyId == "" {
		return nil, errors.New("invalid plex credentials")
	}

	return &Credentials{
		ClientId: clientId,
		KeyId:    keyId,
		Key:      ed25519.NewKeyFromSeed(seed),
	}, nil
}

// plex treats the kid as the key's thumbprint, and a thumbprint can only belong to one device
func thumbprint(pub ed25519.PublicKey) string {
	sum := sha256.Sum256([]byte(`{"crv":"Ed25519","kty":"OKP","x":"` + base64.RawURLEncoding.EncodeToString(pub) + `"}`))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func (c *AuthClient) CreatePin(ctx context.Context, credentials *Credentials) (*Pin, error) {
	b, err := json.Marshal(createPinBody{
		Jwk: jwk{
			Kty: "OKP",
			Crv: "Ed25519",
			X:   base64.RawURLEncoding.EncodeToString(credentials.Key.Public().(ed25519.PublicKey)),
			Kid: credentials.KeyId,
			Alg: "EdDSA",
		},
		Strong: true,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", Url.JoinPath("api", "v2", "pins").String(), bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	return c.executeRequest[Pin](req, credentials.ClientId)
}

func (c *AuthClient) ExchangePin(ctx context.Context, credentials *Credentials, pinId int) (*Pin, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.MapClaims{
		"aud": "plex.tv",
		"iss": credentials.ClientId,
		"iat": now.Unix(),
		"exp": now.Add(time.Minute * 5).Unix(),
	})
	token.Header["kid"] = credentials.KeyId
	deviceJwt, err := token.SignedString(credentials.Key)
	if err != nil {
		return nil, err
	}

	uri := Url.JoinPath("api", "v2", "pins", strconv.Itoa(pinId))
	query := uri.Query()
	query.Set("deviceJWT", deviceJwt)
	uri.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, "GET", uri.String(), nil)
	if err != nil {
		return nil, err
	}

	return c.executeRequest[Pin](req, credentials.ClientId)
}

func (c *AuthClient) GetNonce(ctx context.Context, credentials *Credentials) (*Nonce, error) {
	uri := Url.JoinPath("api", "v2", "auth", "nonce")
	req, err := http.NewRequestWithContext(ctx, "GET", uri.String(), nil)
	if err != nil {
		return nil, err
	}

	return c.executeRequest[Nonce](req, credentials.ClientId)
}

func (c *AuthClient) ExchangePlexToken(ctx context.Context, credentials *Credentials, nonce *Nonce) (*ExchangedAuthToken, error) {
	uri := Url.JoinPath("api", "v2", "auth", "token")
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.MapClaims{
		"nonce": nonce.Value,
		"iss":   credentials.ClientId,
		"iat":   now.Unix(),
		"exp":   now.Add(time.Minute * 5).Unix(),
		"aud":   "plex.tv",
		"scope": "username,email,friendly_name",
	})
	token.Header["kid"] = credentials.KeyId
	token.Header["alg"] = "EdDSA"
	token.Header["typ"] = "JWT"

	tokenString, err := token.SignedString(credentials.Key)
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

	return c.executeRequest[ExchangedAuthToken](req, credentials.ClientId)
}

func (c *AuthClient) SignOut(ctx context.Context, clientId, token string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", Url.JoinPath("api", "v2", "users", "signout").String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Plex-Token", token)

	_, err = c.send(req, clientId)
	return err
}

func (c *AuthClient) executeRequest[T any](req *http.Request, clientId string) (*T, error) {
	body, err := c.send(req, clientId)
	if err != nil {
		return nil, err
	}

	t := new(T)
	if err := json.Unmarshal(body, t); err != nil {
		return nil, err
	}

	return t, nil
}

func (c *AuthClient) send(req *http.Request, clientId string) ([]byte, error) {
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-Plex-Client-Identifier", clientId)
	req.Header.Add("X-Plex-Product", "AutoVolume")

	rsp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			slog.Error("failed to close response body", slog.Any("error", err))
		}
	}(rsp.Body)

	if rsp.StatusCode < http.StatusOK || rsp.StatusCode >= 300 {
		return nil, fmt.Errorf("invalid status code: %d", rsp.StatusCode)
	}

	return io.ReadAll(rsp.Body)
}

func mustParse(u string) *url.URL {
	parsed, err := url.Parse(u)
	if err != nil {
		panic(err)
	}

	return parsed
}
