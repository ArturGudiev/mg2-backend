package models

import "arturgudiev/memoryguard/ent/schema"

type LoginUserRequest struct {
	Email    string `json:"email" binding:"required" example:"john.doe@example.com"`
	Password string `json:"password" binding:"required" example:"password"`
}

type LoginUserResponse struct {
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
	User         UserResponse `json:"user"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type UserResponse struct {
	ID    int             `json:"id"`
	Name  string          `json:"name"`
	Email string          `json:"email"`
	Role  schema.UserRole `json:"role"`
}

type NewUserRequest struct {
	Name     string `json:"name" binding:"required" example:"John Doe"`
	Email    string `json:"email" binding:"required" example:"john.doe@example.com"`
	Password string `json:"password" binding:"required" example:"password"`
}
