package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gr4vediggr/stellarlight/internal/users"
	"github.com/labstack/gommon/log"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo      users.UserRepository
	jwtSecret string
}

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

func NewService(userRepo users.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{repo: userRepo, jwtSecret: jwtSecret}
}

func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	// Check if exists
	_, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return nil, errors.New("user already exists")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.CreateUser(ctx, &users.User{
		ID:          uuid.New(),
		Email:       req.Email,
		DisplayName: req.DisplayName,
		Password:    string(hashed),
	})
	if err != nil {
		return nil, err
	}

	token, refresh, err := s.generateTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token:        token,
		RefreshToken: refresh,
		User:         user,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidUsernameOrPassword
	}

	if err := user.CheckPassword(req.Password); err != nil {
		return nil, ErrInvalidUsernameOrPassword
	}

	token, refreshToken, err := s.generateTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

func (s *AuthService) UpdateProfile(ctx context.Context, userID uuid.UUID, req *UpdateProfileRequest) (*users.User, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {

		return nil, ErrUserNotFound
	}

	// If changing password, verify current password
	if req.NewPassword != "" {
		if req.CurrentPassword == "" {
			return user, ErrCurrentPasswordRequired
		}
		if err := user.CheckPassword(req.CurrentPassword); err != nil {
			return user, ErrCurrentPasswordIncorrect
		}
		if err := user.SetPassword(req.NewPassword); err != nil {
			return user, err
		}
	}

	user.DisplayName = req.DisplayName
	_, err = s.repo.UpdateUser(ctx, user)
	return user, err
}

func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (s *AuthService) generateTokens(ctx context.Context, user *users.User) (string, string, error) {
	now := time.Now()

	// Access token (short-lived)
	accessClaims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", "", err
	}

	// Refresh token (7 days)
	refreshID := uuid.New()
	refreshClaims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        refreshID.String(), // JTI
			ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", "", err
	}

	// Store refresh token in DB
	expiresAt := now.Add(7 * 24 * time.Hour)
	_, err = s.repo.CreateRefreshToken(ctx, &users.CreateRefreshTokenParams{
		ID:        refreshID,
		UserID:    user.ID,
		Token:     refreshTokenString,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshTokenString, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	// Parse and validate
	token, err := jwt.ParseWithClaims(refreshToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("could not parse claims")
	}

	// Lookup in DB
	dbToken, err := s.repo.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, errors.New("refresh token not found")
	}

	if dbToken.Revoked || dbToken.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("refresh token expired or revoked")
	}

	user, err := s.repo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Revoke old refresh token
	err = s.repo.RevokeRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	// Issue new tokens
	accessToken, newRefreshToken, err := s.generateTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token:        accessToken,
		RefreshToken: newRefreshToken,
		User:         user,
	}, nil
}

func (s *AuthService) BackgroundCleanup(ctx context.Context) {
	// Periodically delete expired refresh tokens
	defer log.Info("AuthService: Background cleanup stopped")
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.repo.DeleteExpiredRefreshTokens(ctx); err != nil {
				// Log error (not implemented here)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (h *AuthService) RevokeRefreshToken(ctx context.Context, token string) error {

	if err := h.repo.RevokeRefreshToken(ctx, token); err != nil {
		return err
	}

	return nil
}
