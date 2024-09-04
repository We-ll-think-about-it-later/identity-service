package model

import (
	"crypto/rand"
	"encoding/base64"
	"errors"

	"github.com/We-ll-think-about-it-later/identity-service/pkg/jwt"
)

var (
	ErrEmptyRefreshToken = errors.New("refresh_token can't be empty")
)

// can't use `type Name string`
// due to go's implicit casting
type RefreshToken struct{ value string }

func (t RefreshToken) String() string { return t.value }
func (t RefreshToken) Bytes() []byte  { return []byte(t.value) }

func NewRefreshToken() RefreshToken {
	bs := make([]byte, 48)
	_, _ = rand.Read(bs)
	refreshString := base64.RawURLEncoding.EncodeToString(bs)
	return RefreshToken{value: refreshString}
}

func RefreshTokenFromString(s string) (RefreshToken, error) {
	if s == "" {
		return RefreshToken{}, ErrEmptyRefreshToken
	}

	return RefreshToken{value: s}, nil
}

type AccessToken struct{ jwt.Jwt }

func NewAccessToken(payload AccessTokenPayload, secret []byte) (AccessToken, error) {

	accessJwt, err := jwt.Sign(payload, secret)
	if err != nil {
		return AccessToken{}, err
	}
	return AccessToken{accessJwt}, nil
}

type AccessTokenPayload struct {
	Sub string `json:"sub"`
	Iat int64  `json:"iat"`
	Exp int64  `json:"exp"`
}
