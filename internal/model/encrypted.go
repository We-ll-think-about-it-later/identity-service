package model

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var ErrEmptyString = errors.New("empty string can't be encrypted")

type Encryptable interface {
	Bytes() []byte
}

type Encrypted[T Encryptable] []byte

func (e Encrypted[T]) IsEqual(unencrypted T) bool {
	return bcrypt.CompareHashAndPassword(e, unencrypted.Bytes()) == nil
}
func (e Encrypted[T]) String() string {
	return string(e)
}

func Encrypt[T Encryptable](unencrypted T) (Encrypted[T], error) {
	b := unencrypted.Bytes()
	if len(b) == 0 {
		return nil, ErrEmptyString
	}
	hashed, err := bcrypt.GenerateFromPassword(b, bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return Encrypted[T](hashed), nil
}

func EncryptedFromString[T Encryptable](s string) (Encrypted[T], error) {
	if s == "" {
		return Encrypted[T]{}, ErrEmptyString
	}
	return Encrypted[T]([]byte(s)), nil
}
