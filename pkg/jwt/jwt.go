package jwt

import (
	"crypto/hmac"
	sha256 "crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

type Jwt struct {
	value string
}

func (j Jwt) String() string {
	return j.value
}

var header = []byte(`{"typ": "JWT", "alg": "HS256"}`)

func Sign(payload any, secret []byte) (Jwt, error) {
	p, err := json.Marshal(payload)
	if err != nil {
		return Jwt{}, err
	}

	encodedHeader := base64.RawURLEncoding.EncodeToString(header)
	encodedPayload := base64.RawURLEncoding.EncodeToString(p)

	unsignedToken := fmt.Sprintf("%s.%s", encodedHeader, encodedPayload)
	signature := hmac256Encode(unsignedToken, secret)

	encodedSignature := base64.RawURLEncoding.EncodeToString(signature)

	return Jwt{
		value: fmt.Sprintf("%s.%s", unsignedToken, encodedSignature),
	}, nil
}

func (j Jwt) HasValidSignature(secret []byte) bool {
	lastDot := strings.LastIndex(j.value, ".")
	encodedSignature := j.value[lastDot+1:]
	gotSignature, err := base64.RawURLEncoding.DecodeString(encodedSignature)
	if err != nil {
		return false
	}

	unsignedToken := j.value[:lastDot]

	expectedSignature := hmac256Encode(unsignedToken, secret)
	return hmac.Equal(gotSignature, expectedSignature)
}

func hmac256Encode(unsignedToken string, secret []byte) []byte {
	hash := hmac.New(sha256.New, secret)
	hash.Write([]byte(unsignedToken))
	return hash.Sum(nil)
}
