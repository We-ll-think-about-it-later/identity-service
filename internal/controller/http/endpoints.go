package http

import (
	"errors"
	"net/http"

	// Swagger docs.
	_ "github.com/We-ll-think-about-it-later/identity-service/docs"

	"github.com/We-ll-think-about-it-later/identity-service/internal/controller/http/types"
	"github.com/We-ll-think-about-it-later/identity-service/internal/model"
	"github.com/We-ll-think-about-it-later/identity-service/internal/service"
	"github.com/We-ll-think-about-it-later/identity-service/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

/*
== Endpoints ==

Handlers
  - Signup
  - Login
  - GetTokens
  - Refresh

Errors:
	* if there is an error in the input data, the handler returns HTTP 400,
	* if the user is not found, the handler returns HTTP 404,
	* if there is an internal server error, the handler returns HTTP 500,
	* if the refresh token is invalid, the handler returns HTTP 401,
*/

// @title           Identity service
// @version         1.0

// errorResponse writes an error response to the client.
func errorResponse(c *gin.Context, logger *logger.Logger, err error, statusCode int) {
	logger.Debug(err)
	c.JSON(statusCode, types.NewErrorResponseBody(err))
}

// Signup godoc
// @Summary      Signup
// @Description  Creates a new user.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      types.SignupRequestBody  true  "Signup request body"
// @Success      200  {object}  types.SignupResponseBody
// @Failure      400  {object}  types.ErrorResponseBody
// @Failure      500  {object}  types.ErrorResponseBody
// @Router       /auth/signup [post]
func (s *Server) Signup(c *gin.Context) {
	ctx := c.Request.Context()

	var signupRequestBody types.SignupRequestBody

	err := c.BindJSON(&signupRequestBody)
	if err != nil {
		errorResponse(c, s.Logger, err, http.StatusBadRequest)
		return
	}

	newProfileInfo, err := model.NewProfileInfo(
		signupRequestBody.FirstName,
		signupRequestBody.LastName,
		signupRequestBody.Email,
		signupRequestBody.DeviceFingerprint,
	)
	if err != nil {
		errorResponse(c, s.Logger, err, http.StatusBadRequest)
		return
	}

	user, err := s.Usecase.CreateUser(ctx, newProfileInfo)
	if err != nil {
		errorResponse(c, s.Logger, err, http.StatusInternalServerError)
		return
	}

	err = s.Usecase.SendCode(ctx, user)
	if err != nil {
		s.Logger.Debug(err)
		errorResponse(c, s.Logger, service.ErrFailedToSendCode, http.StatusInternalServerError)
	}

	body := types.NewSignupResponseBody(user.UserId)
	c.JSON(http.StatusOK, body)
}

// Login godoc
// @Summary      Login
// @Description  Logs in a user.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      types.LoginRequestBody  true  "Login request body"
// @Success      200  {object}  types.LoginResponseBody
// @Failure      400  {object}  types.ErrorResponseBody
// @Failure      404  {object}  types.ErrorResponseBody
// @Failure      500  {object}  types.ErrorResponseBody
// @Router       /auth/login [post]
func (s *Server) Login(c *gin.Context) {
	ctx := c.Request.Context()

	var loginRequestBody types.LoginRequestBody
	err := c.BindJSON(&loginRequestBody)
	if err != nil {
		errorResponse(c, s.Logger, err, http.StatusBadRequest)
		return
	}

	user, err := s.Usecase.FindUserByEmail(ctx, loginRequestBody.Email)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			errorResponse(c, s.Logger, err, http.StatusNotFound)
		} else {
			errorResponse(c, s.Logger, err, http.StatusInternalServerError)
		}
		return
	}

	err = s.Usecase.SendCode(ctx, user)
	if err != nil {
		s.Logger.Debug(err)
		errorResponse(c, s.Logger, service.ErrFailedToSendCode, http.StatusInternalServerError)
	}

	body := types.NewLoginResponseBody(user.UserId)
	c.JSON(http.StatusOK, body)
}

// GetTokens godoc
// @Summary      GetTokens
// @Description  Gets access and refresh tokens.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      types.GetTokensRequestBody  true  "Get tokens request body"
// @Success      200  {object}  types.GetTokensResponseBody
// @Failure      400  {object}  types.ErrorResponseBody
// @Failure      404  {object}  types.ErrorResponseBody
// @Failure      500  {object}  types.ErrorResponseBody
// @Router       /auth/get_tokens [post]
func (s *Server) GetTokens(c *gin.Context) {
	ctx := c.Request.Context()

	var getTokensRequestBody types.GetTokensRequestBody
	err := c.BindJSON(&getTokensRequestBody)
	if err != nil {
		errorResponse(c, s.Logger, err, http.StatusBadRequest)
		return
	}

	userId, err := uuid.Parse(getTokensRequestBody.UserID)
	if err != nil {
		errorResponse(c, s.Logger, err, http.StatusBadRequest)
		return
	}

	code, err := model.NewCodeFromInt(getTokensRequestBody.Code)
	if err != nil {
		errorResponse(c, s.Logger, err, http.StatusBadRequest)
		return
	}

	err = s.Usecase.ConfirmUser(ctx, userId, code)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			errorResponse(c, s.Logger, err, http.StatusNotFound)
		} else {
			errorResponse(c, s.Logger, err, http.StatusInternalServerError)
		}
		return
	}

	access, refresh, err := s.Usecase.GetTokens(ctx, userId)
	if err != nil {
		errorResponse(c, s.Logger, err, http.StatusInternalServerError)
		return
	}

	body := types.NewGetTokensResponseBody(access, refresh)
	c.JSON(http.StatusOK, body)
}

// Refresh godoc
// @Summary      Refresh
// @Description  Refreshes access token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      types.RefreshRequestBody  true  "Refresh request body"
// @Success      200  {object}  types.RefreshResponseBody
// @Failure      400  {object}  types.ErrorResponseBody
// @Failure      401  {object}  types.ErrorResponseBody
// @Failure      404  {object}  types.ErrorResponseBody
// @Failure      500  {object}  types.ErrorResponseBody
// @Router       /auth/refresh [post]
func (s *Server) Refresh(c *gin.Context) {
	ctx := c.Request.Context()

	var refreshRequestBody types.RefreshRequestBody
	if err := c.BindJSON(&refreshRequestBody); err != nil {
		errorResponse(c, s.Logger, err, http.StatusBadRequest)
		return
	}

	userId, err := uuid.Parse(refreshRequestBody.UserID)
	if err != nil {
		errorResponse(c, s.Logger, err, http.StatusBadRequest)
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
			errorResponse(c, s.Logger, err, http.StatusUnauthorized)
		} else if errors.Is(err, service.ErrUserNotFound) {
			errorResponse(c, s.Logger, err, http.StatusNotFound)
		} else {
			errorResponse(c, s.Logger, err, http.StatusInternalServerError)
		}
		return
	}

	body := types.NewRefreshResponseBody(access)
	c.JSON(http.StatusOK, body)
}
