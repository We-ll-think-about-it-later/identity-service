package service

import (
	"context"
	"errors"
	"fmt"

	. "github.com/We-ll-think-about-it-later/identity-service/internal/model"
	"github.com/We-ll-think-about-it-later/identity-service/internal/repository"
	. "github.com/We-ll-think-about-it-later/identity-service/pkg/email"
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
	ErrProfileDoesNotExist  = errors.New("profile doesn't exist")
	ErrProfileAlreadyExists = errors.New("profile alredy exists")
)

type UserService interface {
	Authenticate(ctx context.Context, email Email) (uuid.UUID, bool, error)
	CheckCode(ctx context.Context, userId uuid.UUID, code Code) error
	GetTokens(ctx context.Context, userId uuid.UUID) (AccessToken, RefreshToken, error)
	GetUserProfile(ctx context.Context, userId uuid.UUID) (ProfileInfo, error)
	Refresh(ctx context.Context, userId uuid.UUID, token RefreshToken) (AccessToken, error)
	SendCode(ctx context.Context, user User) error
	CreateUserProfile(ctx context.Context, userId uuid.UUID, profileInfo ProfileInfo) (ProfileInfo, error)
	UpdateUserProfile(ctx context.Context, userId uuid.UUID, profileInfo ProfileInfoUpdate) (ProfileInfo, error)
}

type UserServiceImpl struct {
	tokenService   TokenService
	emailSender    *EmailSender
	userRepository repository.UserRepository
	codeRepository repository.CodeRepository
	logger         *logger.Logger
}

func NewUserService(
	tokenService TokenService,
	emailSender *EmailSender,
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

func (s UserServiceImpl) Authenticate(ctx context.Context, email Email) (uuid.UUID, bool, error) {
	isNewUser := false
	// user, err := s.userRepository.FindByEmail(ctx, email)
	//
	// if errors.Is(err, repository.ErrUserNotFound) {
	// 	isNewUser = true
	user, err := s.userRepository.CreateUser(ctx, email)
	// }

	if err != nil {
		return uuid.Nil, false, err
	}

	err = s.SendCode(ctx, user)
	if err != nil {
		return uuid.Nil, false, err
	}

	return user.UserId, isNewUser, nil
}

func (s UserServiceImpl) CheckCode(ctx context.Context, userId uuid.UUID, code Code) error {
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

	return s.codeRepository.DeleteConfirmationCode(ctx, userId)
}

func (s UserServiceImpl) GetTokens(ctx context.Context, userId uuid.UUID) (AccessToken, RefreshToken, error) {
	_, err := s.userRepository.FindById(ctx, userId)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return AccessToken{}, RefreshToken{}, ErrUserNotFound
		}
		return AccessToken{}, RefreshToken{}, fmt.Errorf("failed to find user by id: %w", err)
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

func (s UserServiceImpl) GetUserProfile(ctx context.Context, userId uuid.UUID) (ProfileInfo, error) {
	user, err := s.userRepository.FindById(ctx, userId)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ProfileInfo{}, ErrUserNotFound
		}
		return ProfileInfo{}, err
	}
	if user.ProfileInfo == nil {
		return ProfileInfo{}, ErrProfileDoesNotExist
	}

	return *user.ProfileInfo, nil
}

func (s UserServiceImpl) Refresh(ctx context.Context, userId uuid.UUID, token RefreshToken) (AccessToken, error) {
	isValid := s.tokenService.IsValidRefreshToken(userId, token)
	if !isValid {
		return AccessToken{}, ErrInvalidRefreshToken
	}

	_, err := s.userRepository.FindById(ctx, userId)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return AccessToken{}, ErrUserNotFound
		}
		return AccessToken{}, fmt.Errorf("failed to find user by id: %w", err)
	}

	newAccess, err := s.tokenService.GenerateAccessToken(userId)
	if err != nil {
		return AccessToken{}, err
	}

	return newAccess, nil
}

func (s UserServiceImpl) SendCode(ctx context.Context, user User) error {
	code := NewCode()
	encryptedCode, err := Encrypt(code)
	if err != nil {
		return fmt.Errorf("failed to encrypt confirmation code: %w", err)
	}
	err = s.emailSender.Send(user.Email, "Confirmation code: "+code.String())
	if err != nil {
		return fmt.Errorf("failed to send confirmation code: %w", err)
	}

	err = s.codeRepository.SaveConfirmationCode(ctx, user.UserId, encryptedCode)
	if err != nil {
		return fmt.Errorf("failed to save confirmation code: %w", err)
	}
	return nil
}

func (s UserServiceImpl) CreateUserProfile(ctx context.Context, userId uuid.UUID, profileInfo ProfileInfo) (ProfileInfo, error) {
	user, err := s.userRepository.FindById(ctx, userId)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ProfileInfo{}, ErrUserNotFound
		}
		return ProfileInfo{}, err
	}

	if user.ProfileInfo != nil {
		return ProfileInfo{}, ErrProfileAlreadyExists
	}

	err = s.userRepository.CreateProfile(ctx, userId, profileInfo)
	if err != nil {
		return ProfileInfo{}, err
	}

	return profileInfo, nil
}

func (s UserServiceImpl) UpdateUserProfile(ctx context.Context, userId uuid.UUID, profileInfo ProfileInfoUpdate) (ProfileInfo, error) {
	user, err := s.userRepository.FindById(ctx, userId)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ProfileInfo{}, ErrUserNotFound
		}
		return ProfileInfo{}, err
	}

	if user.ProfileInfo == nil {
		return ProfileInfo{}, ErrProfileDoesNotExist
	}
	newProfile, err := s.userRepository.UpdateProfile(ctx, userId, profileInfo)
	if err != nil {
		return ProfileInfo{}, err
	}

	return newProfile, nil
}
