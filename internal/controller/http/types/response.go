package types

import (
	"github.com/We-ll-think-about-it-later/identity-service/internal/model"
	"github.com/google/uuid"
)

type ErrorResponseBody struct {
	Error string `json:"error"`
}

func NewErrorResponseBody(err error) ErrorResponseBody {
	errMsg := "Internal Server Error"
	if err != nil {
		errMsg = err.Error()
	}

	return ErrorResponseBody{
		Error: errMsg,
	}
}

// AuthenticateResponseBody represents the response body for the /auth/authenticate endpoint.
type AuthenticateResponseBody struct {
	UserID uuid.UUID `json:"user_id"`
}

func NewAuthenticateResponseBody(userId uuid.UUID) AuthenticateResponseBody {
	return AuthenticateResponseBody{
		UserID: userId,
	}
}

// GetTokensResponseBody represents the response body for the /auth/token endpoint.
type GetTokensResponseBody struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func NewGetTokensResponseBody(access model.AccessToken, refresh model.RefreshToken) GetTokensResponseBody {
	return GetTokensResponseBody{
		AccessToken:  access.String(),
		RefreshToken: refresh.String(),
	}
}

// RefreshResponseBody represents the response body for the /auth/token/refresh endpoint.
type RefreshResponseBody struct {
	AccessToken string `json:"access_token"`
}

func NewRefreshResponseBody(access model.AccessToken) RefreshResponseBody {
	return RefreshResponseBody{
		AccessToken: access.String(),
	}
}

// UserProfileResponseBody represents the response body for the /users/{user_id}/profile endpoint.
type UserProfileResponseBody struct {
	Firstname string  `json:"firstname"`
	Lastname  *string `json:"lastname"`
	Username  string  `json:"username"`
}

func NewUserProfileResponseBody(profileInfo model.ProfileInfo) UserProfileResponseBody {
	return UserProfileResponseBody{
		Firstname: profileInfo.FirstName,
		Lastname:  profileInfo.LastName,
		Username:  profileInfo.UserName,
	}
}
