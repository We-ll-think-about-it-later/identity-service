package service

import (
	"context"
	"errors"
	"fmt"

	. "github.com/We-ll-think-about-it-later/identity-service/internal/model"
	"github.com/We-ll-think-about-it-later/identity-service/internal/repository"
	"github.com/We-ll-think-about-it-later/identity-service/pkg/email"
	"github.com/We-ll-think-about-it-later/identity-service/pkg/logger"
	"github.com/google/uuid"
)

var (
	ErrCantSaveRefreshToken = errors.New("can't save refresh token")
	ErrInvalidRefreshToken  = errors.New("refresh token is invalid")
	ErrFailedToSendCode     = errors.New("failed to send code")
	ErrCodeMismatch         = errors.New("code mismatch")
	ErrUserNotFound         = errors.New("user not found")
	ErrUserNotConfirmed     = errors.New("user not confirmed")
)

type UserService interface {
	CreateUser(ctx context.Context, profileInfo ProfileInfo) (User, error)
	FindUserByID(ctx context.Context, userId uuid.UUID) (User, error)
	FindUserByEmail(ctx context.Context, email string) (User, error)
	ConfirmUser(ctx context.Context, userId uuid.UUID, code Code) error
	SendCode(ctx context.Context, user User) error
	Login(ctx context.Context, email string) (User, error)
	GetTokens(ctx context.Context, userId uuid.UUID) (AccessToken, RefreshToken, error)
	Refresh(ctx context.Context, userId uuid.UUID, token RefreshToken) (AccessToken, error)
}

type UserServiceImpl struct {
	tokenService   TokenService
	emailSender    *email.EmailSender
	userRepository repository.UserRepository
	codeRepository repository.CodeRepository
	logger         *logger.Logger
}

func NewUserService(
	tokenService TokenService,
	emailSender *email.EmailSender,
	userRepository repository.UserRepository,
	codeRepository repository.CodeRepository,
	logger *logger.Logger,
) UserServiceImpl {
	logger.SetPrefix("service - user ")

	return UserServiceImpl{
		tokenService:   tokenService,
		emailSender:    emailSender,
		userRepository: userRepository,
		codeRepository: codeRepository,
		logger:         logger,
	}
}

func (s UserServiceImpl) CreateUser(ctx context.Context, profileInfo ProfileInfo) (User, error) {
	user, err := s.userRepository.CreateUser(ctx, profileInfo)
	if err != nil {
		return User{}, fmt.Errorf("failed to create user: %w", err)
	}
	return user, nil
}

func (s UserServiceImpl) SendCode(ctx context.Context, user User) error {
	code := NewCode()
	encryptedCode, err := Encrypt(code)
	if err != nil {
		return fmt.Errorf("failed to encrypt confirmation code: %w", err)
	}
	err = s.emailSender.Send(user.ProfileInfo.Email, "Confirmation code: "+code.String())
	if err != nil {
		return fmt.Errorf("failed to send confirmation code: %w", err)
	}

	err = s.codeRepository.SaveConfirmationCode(ctx, user.UserId, encryptedCode)
	if err != nil {
		return fmt.Errorf("failed to save confirmation code: %w", err)
	}
	return nil
}

func (s UserServiceImpl) FindUserByID(ctx context.Context, userId uuid.UUID) (User, error) {
	user, err := s.userRepository.FindByID(ctx, userId)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("failed to find user by id: %w", err)
	}

	return user, nil
}

func (s UserServiceImpl) FindUserByEmail(ctx context.Context, email string) (User, error) {
	user, err := s.userRepository.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("failed to find user by email: %w", err)
	}

	return user, nil
}

func (s UserServiceImpl) ConfirmUser(ctx context.Context, userId uuid.UUID, code Code) error {
	storedCode, err := s.codeRepository.GetConfirmationCode(ctx, userId)
	if err != nil {
		if errors.Is(err, repository.ErrCodeNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to get confirmation code: %w", err)
	}

	if !storedCode.IsEqual(code) {
		return ErrCodeMismatch
	}

	err = s.codeRepository.DeleteConfirmationCode(ctx, userId)
	s.logger.Debug(err)

	err = s.userRepository.ConfirmUser(ctx, userId)
	if err != nil {
		return fmt.Errorf("failed to confirm user: %w", err)
	}

	return nil
}

func (s UserServiceImpl) Login(ctx context.Context, email string) (User, error) {
	user, err := s.userRepository.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("failed to find user by email: %w", err)
	}

	if !user.IsConfirmed {
		return User{}, ErrUserNotConfirmed
	}

	return user, nil
}

func (s UserServiceImpl) GetTokens(ctx context.Context, userId uuid.UUID) (AccessToken, RefreshToken, error) {
	user, err := s.userRepository.FindByID(ctx, userId)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return AccessToken{}, RefreshToken{}, ErrUserNotFound
		}
		return AccessToken{}, RefreshToken{}, fmt.Errorf("failed to find user by id: %w", err)
	}

	if !user.IsConfirmed {
		return AccessToken{}, RefreshToken{}, ErrUserNotConfirmed
	}

	refresh, err := s.tokenService.GenerateRefreshToken(userId)
	if err != nil {
		return AccessToken{}, RefreshToken{}, err
	}

	access, err := s.tokenService.GenerateAccessToken(userId)
	if err != nil {
		return AccessToken{}, RefreshToken{}, err
	}

	return access, refresh, nil
}

func (s UserServiceImpl) Refresh(ctx context.Context, userId uuid.UUID, token RefreshToken) (AccessToken, error) {
	isValid := s.tokenService.IsValidRefreshToken(userId, token)
	if !isValid {
		return AccessToken{}, ErrInvalidRefreshToken
	}

	user, err := s.userRepository.FindByID(ctx, userId)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return AccessToken{}, ErrUserNotFound
		}
		return AccessToken{}, fmt.Errorf("failed to find user by id: %w", err)
	}

	if !user.IsConfirmed {
		return AccessToken{}, ErrUserNotConfirmed
	}

	newAccess, err := s.tokenService.GenerateAccessToken(userId)
	if err != nil {
		return AccessToken{}, err
	}

	return newAccess, nil
}
