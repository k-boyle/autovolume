package plex

import "time"

type Pin struct {
	Id               int    `json:"id"`
	Code             string `json:"code"`
	ClientIdentifier string `json:"clientIdentifier"`
	// utc timestamps
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`

	AuthToken *string `json:"authToken,omitempty"`
}

type Users struct {
	Subscription bool   `json:"subscription"`
	Users        []User `json:"users"`
}

type User struct {
	Username  string `json:"username"`
	AvatarUrl string `json:"thumb"`
}

type Nonce struct {
	Value string `json:"nonce"`
}

type jwk struct {
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
}

type createPinBody struct {
	Jwk    jwk  `json:"jwk"`
	Strong bool `json:"strong"`
}

type exchangeAuthBody struct {
	Jwt string `json:"jwt"`
}
