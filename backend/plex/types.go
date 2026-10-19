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

type Device struct {
	Id          string        `json:"clientIdentifier"`
	Name        string        `json:"name"`
	Provides    string        `json:"provides"`
	Connections []*Connection `json:"connections"`
	AccessToken string        `json:"token"`
	LastSeenAt  time.Time     `json:"lastSeenAt"`
	Product     string        `json:"product"`
	Presence    bool          `json:"presence"`
}

type Connection struct {
	Uri string `json:"uri"`
}

type Identity struct {
}

type LibrarySections struct {
	Directory []LibraryDirectory `json:"Directory"`
}

type LibraryDirectory struct {
	Key   string `json:"key"`
	Title string `json:"title"`
	Type  string `json:"type"`
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

type mediaContainer[T any] struct {
	Value T `json:"mediaContainer"`
}

type ExchangedAuthToken struct {
	AuthToken string `json:"auth_token"`
}
