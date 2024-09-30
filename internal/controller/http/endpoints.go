package http

// @title           Identity service
// @version         1.0

import (
	"errors"
	"net/http"

	// Swagger docs.
	_ "github.com/We-ll-think-about-it-later/identity-service/docs"

	"github.com/We-ll-think-about-it-later/identity-service/internal/controller/http/types"
	"github.com/We-ll-think-about-it-later/identity-service/internal/model"
	"github.com/We-ll-think-about-it-later/identity-service/internal/service"
	"github.com/We-ll-think-about-it-later/identity-service/pkg/email"
	"github.com/We-ll-think-about-it-later/identity-service/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// errorResponse writes an error response to the client.
func errorResponse(c *gin.Context, logger *logger.Logger, err error, statusCode int) {
	logger.Debug(err)
	c.JSON(statusCode, types.NewErrorResponseBody(err))
}

// Authenticate godoc
// @Summary      Authenticate
// @Description  Authenticates a user and sends a verification code to their email. Returns the user's UUID.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      types.AuthenticateRequestBody  true  "Authentication request body"
// @Param        X-Device-Fingerprint   header     string  true  "SHA-256 hash of device fingerprint"
// @Success      200  {object}  types.AuthenticateResponseBody  "User found"
// @Success      201  {object}  types.AuthenticateResponseBody  "User created"
// @Failure      400  {object}  types.ErrorResponseBody  "Invalid request"
// @Router       /auth/authenticate [post]
func (s *Server) Authenticate(c *gin.Context) {
	ctx := c.Request.Context()

	var authenticateRequestBody types.AuthenticateRequestBody

	err := c.BindJSON(&authenticateRequestBody)
	if err != nil {
		errorResponse(c, s.Logger, err, http.StatusBadRequest)
		return
	}

	email, err := email.NewEmail(authenticateRequestBody.Email)
	if err != nil {
		errorResponse(c, s.Logger, err, http.StatusBadRequest)
		return
	}

	userId, isNewUser, err := s.Usecase.Authenticate(ctx, email)
	if err != nil {
		s.Logger.Error(err)
		errorResponse(c, s.Logger, nil, http.StatusInternalServerError)
		return
	}

	body := types.NewAuthenticateResponseBody(userId)

	if isNewUser {
		c.JSON(http.StatusCreated, body)
	} else {
		c.JSON(http.StatusOK, body)
	}
}

// GetTokens godoc
// @Summary      GetTokens
// @Description  Retrieves access and refresh tokens after verification code entry.  The refresh token is set in an HTTP-only cookie by the API Gateway.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      types.GetTokensRequestBody  true  "Get tokens request body"
// @Param        X-User-Id              header     string  true  "User ID"
// @Param        X-Device-Fingerprint   header     string  true  "SHA-256 hash of device fingerprint"
// @Success      200  {object}  types.GetTokensResponseBody  "Tokens issued"
// @Failure      400  {object}  types.ErrorResponseBody  "Invalid request"
// @Failure      401  {object}  types.ErrorResponseBody  "Invalid user ID or user not found"
// @Failure      403  {object}  types.ErrorResponseBody "Invalid verification code"
// @Failure      500  {object}  types.ErrorResponseBody "Internal server error"
// @Router       /auth/token [post]
func (s *Server) GetTokens(c *gin.Context) {
	ctx := c.Request.Context()

	var getTokensRequestBody types.GetTokensRequestBody
	err := c.BindJSON(&getTokensRequestBody)
	if err != nil {
		errorResponse(c, s.Logger, err, http.StatusBadRequest)
		return
	}

	userId, err := uuid.Parse(c.GetHeader("X-User-Id"))
	if err != nil {
		errorResponse(c, s.Logger, err, http.StatusBadRequest)
		return
	}

	code, err := model.NewCodeFromInt(getTokensRequestBody.Code)
	if err != nil {
		errorResponse(c, s.Logger, err, http.StatusUnauthorized)
		return
	}

	err = s.Usecase.CheckCode(ctx, userId, code)
	if err != nil {
		if errors.Is(err, service.ErrCodeMismatch) {
			errorResponse(c, s.Logger, err, http.StatusUnauthorized)
			return
		}
		if errors.Is(err, service.ErrUserNotFound) {
			errorResponse(c, s.Logger, err, http.StatusUnauthorized)
			return
		}
		s.Logger.Error(err)
		errorResponse(c, s.Logger, nil, http.StatusInternalServerError)
		return
	}

	access, refresh, err := s.Usecase.GetTokens(ctx, userId)
	if err != nil {
		s.Logger.Error(err)
		errorResponse(c, s.Logger, nil, http.StatusInternalServerError)
		return
	}

	body := types.NewGetTokensResponseBody(access, refresh)
	c.JSON(http.StatusOK, body)
}

// Refresh godoc
// @Summary      Refresh
// @Description  Refreshes the access token. The refresh token is retrieved from an HTTP-only cookie by the API Gateway.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      types.RefreshRequestBody  true  "Refresh request body"
// @Param        X-User-Id              header     string  true  "User ID"
// @Param        X-Device-Fingerprint   header     string  true  "SHA-256 hash of device fingerprint"
// @Success      200  {object}  types.RefreshResponseBody  "Access token refreshed"
// @Failure      400  {object}  types.ErrorResponseBody  "Invalid request"
// @Failure      401  {object}  types.ErrorResponseBody  "Invalid user ID or user not found"
// @Failure      403  {object}  types.ErrorResponseBody  "Invalid refresh token"
// @Failure      404  {object}  types.ErrorResponseBody "User not found"
// @Failure      500  {object}  types.ErrorResponseBody "Internal server error"
// @Router       /auth/token/refresh [post]
func (s *Server) Refresh(c *gin.Context) {
	ctx := c.Request.Context()

	var refreshRequestBody types.RefreshRequestBody
	if err := c.BindJSON(&refreshRequestBody); err != nil {
		errorResponse(c, s.Logger, err, http.StatusBadRequest)
		return
	}

	userId, err := uuid.Parse(c.GetHeader("X-User-Id"))
	if err != nil {
		errorResponse(c, s.Logger, err, http.StatusUnauthorized)
		return
	}

	refreshToken, err := model.RefreshTokenFromString(refreshRequestBody.RefreshToken)
	if err != nil {
		errorResponse(c, s.Logger, err, http.StatusBadRequest)
		return
	}

	access, err := s.Usecase.Refresh(ctx, userId, refreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidRefreshToken) {
			errorResponse(c, s.Logger, err, http.StatusForbidden)
		} else if errors.Is(err, service.ErrUserNotFound) {
			errorResponse(c, s.Logger, err, http.StatusNotFound)
		} else {
			s.Logger.Error(err)
			errorResponse(c, s.Logger, nil, http.StatusInternalServerError)
		}
		return
	}

	body := types.NewRefreshResponseBody(access)
	c.JSON(http.StatusOK, body)
}

// CreateUserProfile godoc
// @Summary      CreateUserProfile
// @Description  Creates user profile information.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user_id   path      string  true  "User ID"
// @Param        input  body      types.CreateUserProfileRequestBody  true  "Create user profile request body"
// @Param        X-Device-Fingerprint   header     string  true  "SHA-256 hash of device fingerprint"
// @Success      201  {object}  types.UserProfileResponseBody  "Profile created"
// @Failure      400  {object}  types.ErrorResponseBody  "Invalid request"
// @Failure      404  {object}  types.ErrorResponseBody  "User not found"
// @Failure      409  {object}  types.ErrorResponseBody "Username already taken" // Added 409 conflict
// @Failure      500  {object}  types.ErrorResponseBody "Internal server error"
// @Router       /users/{user_id}/profile [post]
func (s *Server) CreateUserProfile(c *gin.Context) {
	ctx := c.Request.Context()

	userId, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		errorResponse(c, s.Logger, err, http.StatusBadRequest)
		return
	}

	var createUserProfileRequestBody types.CreateUserProfileRequestBody
	if err := c.BindJSON(&createUserProfileRequestBody); err != nil {
		errorResponse(c, s.Logger, err, http.StatusBadRequest)
		return
	}

	profile := createUserProfileRequestBody.ToProfileInfo()
	newProfile, err := s.Usecase.CreateUserProfile(ctx, userId, profile)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			errorResponse(c, s.Logger, err, http.StatusNotFound)
		} else {
			s.Logger.Error(err)
			errorResponse(c, s.Logger, nil, http.StatusInternalServerError)
		}
		return
	}

	body := types.NewUserProfileResponseBody(newProfile)
	c.JSON(http.StatusCreated, body)
}

// UpdateUserProfile godoc
// @Summary      UpdateUserProfile
// @Description  Updates user profile information.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user_id   path      string  true  "User ID"
// @Param        input  body      types.UpdateUserProfileRequestBody  true  "Update user profile request body"
// @Param        X-Device-Fingerprint   header     string  true  "SHA-256 hash of device fingerprint"
// @Success      200  {object}  types.UserProfileResponseBody  "Profile updated"
// @Failure      400  {object}  types.ErrorResponseBody  "Invalid request"
// @Failure      404  {object}  types.ErrorResponseBody  "User or profile not found"
// @Failure      409  {object}  types.ErrorResponseBody "Username already taken" // Added 409 conflict
// @Failure      500  {object}  types.ErrorResponseBody "Internal server error"
// @Router       /users/{user_id}/profile [patch]
func (s *Server) UpdateUserProfile(c *gin.Context) {
	ctx := c.Request.Context()

	userId, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		errorResponse(c, s.Logger, err, http.StatusBadRequest)
		return
	}

	var updateUserProfileRequestBody types.UpdateUserProfileRequestBody
	if err := c.BindJSON(&updateUserProfileRequestBody); err != nil {
		errorResponse(c, s.Logger, err, http.StatusBadRequest)
		return
	}

	profileInfoUpdate := updateUserProfileRequestBody.ToProfileInfoUpdate()
	newProfile, err := s.Usecase.UpdateUserProfile(ctx, userId, profileInfoUpdate)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			errorResponse(c, s.Logger, err, http.StatusNotFound)
		} else {
			s.Logger.Error(err)
			errorResponse(c, s.Logger, nil, http.StatusInternalServerError)
		}
		return
	}

	body := types.NewUserProfileResponseBody(newProfile)
	c.JSON(http.StatusOK, body)
}

// GetUserProfile godoc
// @Summary      GetUserProfile
// @Description  Gets user profile information.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user_id   path      string  true  "User ID"
// @Param        X-Device-Fingerprint   header     string  true  "SHA-256 hash of device fingerprint"
// @Success      200  {object}  types.UserProfileResponseBody  "Profile found"
// @Failure      400  {object}  types.ErrorResponseBody  "Invalid request"
// @Failure      404  {object}  types.ErrorResponseBody  "User or profile not found"
// @Failure      500  {object}  types.ErrorResponseBody "Internal server error"
// @Router       /users/{user_id}/profile [get]
func (s *Server) GetUserProfile(c *gin.Context) {
	ctx := c.Request.Context()

	userId, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		errorResponse(c, s.Logger, err, http.StatusBadRequest)
		return
	}

	userProfile, err := s.Usecase.GetUserProfile(ctx, userId)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			errorResponse(c, s.Logger, err, http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrProfileDoesNotExist) {
			errorResponse(c, s.Logger, err, http.StatusNotFound)
			return
		}
		s.Logger.Error(err)
		errorResponse(c, s.Logger, nil, http.StatusInternalServerError)
		return
	}

	body := types.NewUserProfileResponseBody(userProfile)
	c.JSON(http.StatusOK, body)
}
