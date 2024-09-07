package types

import "github.com/google/uuid"

// SignupRequestBody represents the request body for the /auth/signup endpoint.
type SignupRequestBody struct {
	FirstName         string `json:"firstname" binding:"required"`
	LastName          string `json:"lastname"`
	Email             string `json:"email" binding:"required"`
	DeviceFingerprint string `json:"device_fingerprint" binding:"required"`
}

// LoginRequestBody represents the request body for the /auth/login endpoint.
type LoginRequestBody struct {
	Email             string `json:"email" binding:"required"`
	DeviceFingerprint string `json:"device_fingerprint" binding:"required"`
}

// GetTokensRequestBody represents the request body for the /auth/get_tokens endpoint.
type GetTokensRequestBody struct {
	UserID string `json:"user_id" binding:"required,uuid"`
	Code   int    `json:"code" binding:"required"`
}

// RefreshRequestBody represents the request body for the /auth/refresh endpoint.
type RefreshRequestBody struct {
	UserID       string `json:"user_id" binding:"required,uuid"`
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// UserMeRequestBody represents the request body for the /auth/refresh endpoint.
type UserMeRequestBody struct {
	UserID uuid.UUID `json:"user_id" binding:"required,uuid"`
}
