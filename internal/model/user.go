package model

import (
	"errors"

	"github.com/We-ll-think-about-it-later/identity-service/pkg/email"
	"github.com/google/uuid"
)

var (
	ErrInvalidEmail = errors.New("")
)

type ProfileInfo struct {
	FirstName         string `bson:"firstname"`
	LastName          string `bson:"lastname"`
	Email             string `bson:"email"`
	DeviceFingerprint string `bson:"device_fingerprint"`
}

type User struct {
	UserId      uuid.UUID   `bson:"_id"`
	IsConfirmed bool        `bson:"is_confirmed"`
	ProfileInfo ProfileInfo `bson:"profile_info"`
}

func NewProfileInfo(firstname, lastname, mail, deviceFingerprint string) (ProfileInfo, error) {
	if !email.IsValidEmail(mail) {
		return ProfileInfo{}, ErrInvalidEmail
	}
	return ProfileInfo{
		FirstName:         firstname,
		LastName:          lastname,
		Email:             mail,
		DeviceFingerprint: deviceFingerprint,
	}, nil
}

func NewUser(profile ProfileInfo) User {
	userId := uuid.New()
	return User{UserId: userId, ProfileInfo: profile, IsConfirmed: false}
}
