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

type LibrarySection struct {
	TotalSize int               `json:"totalSize"`
	Metadata  []LibraryMetadata `json:"Metadata"`
}

type GetLibrarySectionParams struct {
	Id    string
	Type  int    `query:"type,omitempty"`
	Title string `query:"title,omitempty"`
	Start int    `query:"X-Plex-Container-Start"`
	Size  int    `query:"X-Plex-Container-Size,omitempty"`
}

type GetLibraryChildrenParams struct {
	RatingKey string
}

type LibraryChildren struct {
	ViewGroup string            `json:"viewGroup"`
	Metadata  []LibraryMetadata `json:"Metadata"`
}

type GetLibraryMetadataParams struct {
	RatingKey string
}

// i really hate the plex api
type LibraryMetadatas struct {
	Metadata []LibraryMetadata `json:"Metadata"`
}

type LibraryMetadata struct {
	RatingKey            string `json:"ratingKey"`
	Type                 string `json:"type"`
	Title                string `json:"title"`
	Year                 int    `json:"year"`
	ThumbUrl             string `json:"thumb"`
	ParentTitle          string `json:"parentTitle"`
	Index                int    `json:"index"`
	ParentThumbUrl       string `json:"parentThumb"`
	ParentRatingKey      string `json:"parentRatingKey"`
	ParentIndex          int    `json:"parentIndex"`
	GrandparentRatingKey string `json:"grandparentRatingKey"`
	GrandparentTitle     string `json:"grandparentTitle"`
	GrandparentThumbUrl  string `json:"grandparentThumb"`
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

type websocketMessage struct {
	NotificationContainer notificationContainer `json:"NotificationContainer"`
}

type notificationContainer struct {
	Type                         string                          `json:"type"`
	PlaySessionStateNotification []*PlaySessionStateNotification `json:"PlaySessionStateNotification"`
}

type PlaySessionStateNotification struct {
	ClientIdentifier string `json:"clientIdentifier"`
	RatingKey        string `json:"ratingKey"`
	State            string `json:"state"`
}
