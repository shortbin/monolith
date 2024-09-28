package dto

import (
	"time"
)

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type RegisterReq struct {
	Email    string `json:"email"  validate:"required,email"`
	Password string `json:"password"  validate:"required,password"`
}

type RegisterRes struct {
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginReq struct {
	Email    string `json:"email"  validate:"required,email"`
	Password string `json:"password"  validate:"required"`
}

type LoginRes struct {
	User         User   `json:"user_details"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenReq struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshTokenRes struct {
	AccessToken string `json:"access_token"`
}

type ChangePasswordReq struct {
	Password    string `json:"password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required"`
}

type ForgotPasswordReq struct {
	Email string `json:"email"  validate:"required,email"`
}

type ResetPasswordReq struct {
	ResetToken string `json:"reset_token" validate:"required"`
	Password   string `json:"password" validate:"required"`
}
