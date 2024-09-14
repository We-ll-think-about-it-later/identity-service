package types

import "github.com/We-ll-think-about-it-later/identity-service/internal/model"

// AuthenticateRequestBody represents the request body for the /auth/authenticate endpoint.
type AuthenticateRequestBody struct {
	Email string `json:"email" binding:"required"`
}

// GetTokensRequestBody represents the request body for the /auth/token endpoint.
type GetTokensRequestBody struct {
	Code int `json:"code" binding:"required"`
}

// RefreshRequestBody represents the request body for the /auth/token/refresh endpoint.
type RefreshRequestBody struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// CreateUserProfileRequestBody represents the request body for the /users/{user_id}/profile endpoint (POST).
type CreateUserProfileRequestBody struct {
	Username  string  `json:"username" binding:"required"`
	Firstname string  `json:"firstname" binding:"required"`
	Lastname  *string `json:"lastname"`
}

func (cuprb CreateUserProfileRequestBody) ToProfileInfo() model.ProfileInfo {
	return model.ProfileInfo{
		UserName:  cuprb.Username,
		FirstName: cuprb.Firstname,
		LastName:  cuprb.Lastname,
	}
}

// UpdateUserProfileRequestBody represents the request body for the /users/{user_id}/profile endpoint (PATCH).
type UpdateUserProfileRequestBody struct {
	Username  *string `json:"username"`
	Firstname *string `json:"firstname"`
	Lastname  *string `json:"lastname"`
}

func (uuprb UpdateUserProfileRequestBody) ToProfileInfoUpdate() model.ProfileInfoUpdate {
	return model.ProfileInfoUpdate{
		UserName:  uuprb.Username,
		FirstName: uuprb.Firstname,
		LastName:  uuprb.Lastname,
	}
}
