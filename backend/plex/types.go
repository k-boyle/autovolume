package plex

import "time"

type Pin struct {
	Id               int    `json:"id"`
	Code             string `json:"code"`
	ClientIdentifier string `json:"clientIdentifier"`
	// utc timestamps
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
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
