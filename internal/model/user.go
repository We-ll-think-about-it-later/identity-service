package model

import (
	"github.com/We-ll-think-about-it-later/identity-service/pkg/email"
	"github.com/google/uuid"
)

type ProfileInfo struct {
	FirstName string  `bson:"firstname"`
	LastName  *string `bson:"lastname" validate:"omitempty"`
	UserName  string  `bson:"username"`
}

type ProfileInfoUpdate struct {
	FirstName *string `bson:"firstname" validate:"omitempty"`
	LastName  *string `bson:"lastname" validate:"omitempty"`
	UserName  *string `bson:"username" validate:"omitempty"`
}

type User struct {
	UserId      uuid.UUID    `bson:"_id"`
	Email       email.Email  `bson:"email"`
	ProfileInfo *ProfileInfo `bson:"profile_info" validate:"omitempty"`
}

type UserSearchResult struct {
	UserId      uuid.UUID    `bson:"_id"`
	ProfileInfo *ProfileInfo `bson:"profile_info"`
}

func NewProfileInfo(username, firstname string, lastname *string) ProfileInfo {
	return ProfileInfo{
		FirstName: firstname,
		LastName:  lastname,
		UserName:  username,
	}
}

func NewUser(email email.Email) User {
	userId := uuid.New()
	return User{UserId: userId, Email: email, ProfileInfo: nil}
}
