package service

import (
	"errors"
	"time"

	. "github.com/We-ll-think-about-it-later/identity-service/internal/model"
	"github.com/We-ll-think-about-it-later/identity-service/internal/repository"
	"github.com/We-ll-think-about-it-later/identity-service/pkg/logger"
	"github.com/google/uuid"
)

var (
	ErrRefreshTokenCantBeEncrypted = errors.New("refresh token can't be encrypted")
	ErrCantGenerateAccessToken     = errors.New("can't generate access token")
)

type TokenService interface {
	GenerateAccessToken(uuid.UUID) (AccessToken, error)
	GenerateRefreshToken(uuid.UUID) (RefreshToken, error)
	IsValidRefreshToken(uuid.UUID, RefreshToken) bool
}

type TokenServiceImpl struct {
	accessLifetime int64
	accessSecret   []byte
	repository     repository.TokenRepository
	logger         *logger.Logger
}

func NewTokenService(
	repo repository.TokenRepository,
	accessLifetime int64,
	accessSecret []byte,
	logger *logger.Logger,
) TokenServiceImpl {

	logger.SetPrefix("service - token ")

	return TokenServiceImpl{
		accessLifetime: accessLifetime,
		accessSecret:   accessSecret,
		repository:     repo,
		logger:         logger,
	}
}

func (s TokenServiceImpl) GenerateAccessToken(userId uuid.UUID) (AccessToken, error) {

	now := time.Now().Unix()

	payload := AccessTokenPayload{
		Sub: userId.String(),
		Iat: now,
		Exp: now + s.accessLifetime,
	}

	access, err := NewAccessToken(payload, s.accessSecret)
	if err != nil {
		s.logger.Debug(err)
		return AccessToken{}, ErrCantGenerateAccessToken
	}

	return access, nil
}

func (s TokenServiceImpl) GenerateRefreshToken(userId uuid.UUID) (RefreshToken, error) {
	refreshToken := NewRefreshToken()
	err := s.saveRefreshToken(userId, refreshToken)
	if err != nil {
		return RefreshToken{}, err
	}
	return refreshToken, nil
}

func (s TokenServiceImpl) saveRefreshToken(userId uuid.UUID, token RefreshToken) error {
	encryptedRefresh, err := Encrypt(token)
	if err != nil {
		s.logger.Debug(err)
		return ErrRefreshTokenCantBeEncrypted
	}

	return s.repository.UpsertRefreshToken(userId, encryptedRefresh)
}

func (s TokenServiceImpl) IsValidRefreshToken(userId uuid.UUID, token RefreshToken) bool {
	storedToken, err := s.repository.FindRefreshToken(userId)
	if err != nil {
		s.logger.Debug(err)
		return false
	}

	return storedToken.IsEqual(token)
}
