package mock

import (
	. "github.com/We-ll-think-about-it-later/identity-service/internal/model"
	"github.com/We-ll-think-about-it-later/identity-service/internal/repository"
	"github.com/google/uuid"
)

type TokenRepositoryMock struct {
	storage map[string]string
}

func NewTokenRepositoryMock(userId uuid.UUID, encryptedToken Encrypted[RefreshToken]) TokenRepositoryMock {

	storage := make(map[string]string)
	storage[userId.String()] = encryptedToken.String()

	return TokenRepositoryMock{
		storage: storage,
	}
}

func (r TokenRepositoryMock) UpsertRefreshToken(userId uuid.UUID, refreshToken Encrypted[RefreshToken]) error {
	r.storage[userId.String()] = refreshToken.String()
	return nil
}

func (r TokenRepositoryMock) FindRefreshToken(userId uuid.UUID) (Encrypted[RefreshToken], error) {
	token, ok := r.storage[userId.String()]
	if !ok {
		return Encrypted[RefreshToken]{}, repository.ErrCantFindRefreshToken
	}

	encryptedToken, _ := EncryptedFromString[RefreshToken](token)
	return encryptedToken, nil
}
