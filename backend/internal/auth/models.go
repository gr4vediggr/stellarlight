package auth

import "github.com/gr4vediggr/stellarlight/internal/users"

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type RegisterRequest struct {
	Email       string `json:"email" validate:"required,email"`
	DisplayName string `json:"displayName" validate:"required,min=2,max=50"`
	Password    string `json:"password" validate:"required,min=6"`
}

type UpdateProfileRequest struct {
	DisplayName     string `json:"displayName" validate:"required,min=2,max=50"`
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

type AuthResponse struct {
	Token        string      `json:"token"`
	RefreshToken string      `json:"refreshToken"`
	User         *users.User `json:"user"`
}
