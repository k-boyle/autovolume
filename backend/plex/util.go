package plex

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TokenExpiry(token string) (time.Time, error) {
	claims := jwt.MapClaims{}
	if _, _, err := new(jwt.Parser).ParseUnverified(token, claims); err != nil {
		return time.Time{}, err
	}

	exp, err := claims.GetExpirationTime()
	if err != nil {
		return time.Time{}, err
	}
	if exp == nil {
		return time.Time{}, errors.New("token has no expiration time")
	}

	return exp.Time, nil
}
