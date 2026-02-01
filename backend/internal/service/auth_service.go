package service

import (
	"context"
	"fmt"
	"time"

	"massrouter.ai/backend/internal/model"
	"massrouter.ai/backend/internal/repository"
	"massrouter.ai/backend/pkg/auth"
	"massrouter.ai/backend/pkg/utils"
)

type authService struct {
	userRepo   repository.UserRepository
	jwtManager *auth.JWTManager
}

func NewAuthService(userRepo repository.UserRepository, jwtManager *auth.JWTManager) AuthService {
	return &authService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

func (s *authService) Register(ctx context.Context, req *RegisterRequest) (*model.User, error) {
	existingUser, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("email already registered")
	}

	existingUser, err = s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing username: %w", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("username already taken")
	}

	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &model.User{
		Email:         req.Email,
		Username:      req.Username,
		PasswordHash:  passwordHash,
		Role:          "user",
		Status:        "active",
		EmailVerified: false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *authService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if user.Status != "active" {
		return nil, fmt.Errorf("account is %s", user.Status)
	}

	if !utils.VerifyPassword(req.Password, user.PasswordHash) {
		return nil, fmt.Errorf("invalid credentials")
	}

	tokenPair, err := s.jwtManager.GenerateTokenPair(user.ID, user.Username, user.Email, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("failed to update last login: %w", err)
	}

	return &LoginResponse{
		User:         user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	tokenPair, err := s.jwtManager.RefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return &TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

func (s *authService) Logout(ctx context.Context, userID string) error {
	return nil
}

func (s *authService) VerifyEmail(ctx context.Context, userID string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	if user.EmailVerified {
		return nil
	}

	if err := s.userRepo.Update(ctx, &model.User{
		ID:            user.ID,
		EmailVerified: true,
		UpdatedAt:     time.Now(),
	}); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (s *authService) RequestPasswordReset(ctx context.Context, email string) error {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil
	}

	return nil
}

func (s *authService) ResetPassword(ctx context.Context, token, newPassword string) error {
	return nil
}
