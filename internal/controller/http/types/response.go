package types

import (
	"github.com/We-ll-think-about-it-later/identity-service/internal/model"
	"github.com/google/uuid"
)

type ErrorResponseBody struct {
	Error string `json:"error"`
}

func NewErrorResponseBody(err error) ErrorResponseBody {
	return ErrorResponseBody{
		Error: err.Error(),
	}
}

// SignupResponseBody represents the response body for the /auth/signup endpoint.
type SignupResponseBody struct {
	UserId string `json:"user_id"`
}

func NewSignupResponseBody(userId uuid.UUID) SignupResponseBody {
	return SignupResponseBody{
		UserId: userId.String(),
	}
}

// LoginResponseBody represents the response body for the /auth/login endpoint.
type LoginResponseBody struct {
	UserId string `json:"user_id"`
}

func NewLoginResponseBody(userId uuid.UUID) LoginResponseBody {
	return LoginResponseBody{
		UserId: userId.String(),
	}
}

// GetTokensResponseBody represents the response body for the /auth/get_tokens endpoint.
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

// RefreshResponseBody represents the response body for the /auth/refresh endpoint.
type RefreshResponseBody struct {
	AccessToken string `json:"access_token"`
}

func NewRefreshResponseBody(access model.AccessToken) RefreshResponseBody {
	return RefreshResponseBody{
		AccessToken: access.String(),
	}
}
