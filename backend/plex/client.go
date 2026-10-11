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
)

var ClientsUrl *url.URL
var AppUrl *url.URL

func init() {
	ClientsUrl, _ = url.Parse("https://clients.plex.tv")
	AppUrl, _ = url.Parse("https://app.plex.tv")
}

type Client struct {
	client   *http.Client
	pub      ed25519.PublicKey
	prv      ed25519.PrivateKey
	clientId string
}

func New(client *http.Client, clientId string, pub ed25519.PublicKey, prv ed25519.PrivateKey) (*Client, error) {
	return &Client{
		client:   client,
		pub:      pub,
		prv:      prv,
		clientId: clientId,
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

	req, err := http.NewRequestWithContext(ctx, "POST", ClientsUrl.JoinPath("api", "v2", "pins").String(), bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Plex-Client-Identifier", c.clientId)
	req.Header.Set("X-Plex-Product", "AutoVolume")

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

	if rsp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("invalid status code: %d", rsp.StatusCode)
	}

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}

	pin := &Pin{}
	if err := json.Unmarshal(body, pin); err != nil {
		return nil, err
	}

	return pin, nil
}
