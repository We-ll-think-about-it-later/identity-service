package service

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	. "github.com/We-ll-think-about-it-later/identity-service/internal/model"
	"github.com/We-ll-think-about-it-later/identity-service/internal/repository/mock"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var (
	userId            = uuid.New()
	validToken, _     = RefreshTokenFromString("G65cEalG2Yv9JGLTBvUfwG65cEalG2Yv9JGLTBvUfwd4ZDsqXTky4d4ZDsqXTky4")
	invalidToken, _   = RefreshTokenFromString("JjtUbYN0alR0CujQCK3w3gssWXYkJ8ddV6ohdRf9HQj01g8H7Veq9TZkyd4ZDsqX")
	encryptedToken, _ = Encrypt(validToken)
	lifetime          = int64(120)
	secret            = []byte("secret")
	logger            = logrus.New()
)

func TestGenerateAccessToken(t *testing.T) {
	repo := mock.NewTokenRepositoryMock(userId, encryptedToken)
	tokenService := NewTokenService(repo, lifetime, secret, logger)

	accessToken, err := tokenService.GenerateAccessToken(userId)
	if err != nil {
		t.Errorf("no errors expected, but got %s", err.Error())
	}

	checkAccessToken(t, accessToken, lifetime, secret)
}

func TestIsValidRefreshToken_Valid(t *testing.T) {
	repo := mock.NewTokenRepositoryMock(userId, encryptedToken)
	tokenService := NewTokenService(repo, lifetime, secret, logger)

	isValid := tokenService.IsValidRefreshToken(userId, validToken)
	if !isValid {
		t.Error("expected valid refresh token")
	}
}

func TestIsValidRefreshToken_Invalid(t *testing.T) {
	repo := mock.NewTokenRepositoryMock(userId, encryptedToken)
	tokenService := NewTokenService(repo, lifetime, secret, logger)

	isValid := tokenService.IsValidRefreshToken(userId, invalidToken)
	if isValid {
		t.Error("expected invalid refresh token")
	}
}

func checkAccessToken(t *testing.T, accessToken AccessToken, lifetime int64, secret []byte) {
	jwtParts := strings.Split(accessToken.String(), ".")
	payloadData, _ := base64.RawURLEncoding.DecodeString(jwtParts[1])

	var payload AccessTokenPayload
	json.Unmarshal(payloadData, &payload)

	if payload.Sub != userId.String() {
		t.Errorf("sub mismatch in payload, got != expected: %s != %s", userId.String(), payload.Sub)
	}

	actualLifetime := payload.Exp - payload.Iat
	if actualLifetime != lifetime {
		t.Errorf("lifetime mismatch in payload, got != expected: %d != %d", actualLifetime, lifetime)
	}

	if !accessToken.HasValidSignature(secret) {
		t.Error("expected valid access token signature")
	}
}
