package models

import "arturgudiev/memoryguard/ent/schema"

type LoginUserRequest struct {
	// Login accepts either the user's login or email.
	Login string `json:"login" example:"johndoe"`
	// Email is accepted as an alias for login (backward compatible).
	Email    string `json:"email" example:"john.doe@example.com"`
	Password string `json:"password" binding:"required" example:"password"`
}

// Identifier returns login or email (whichever was provided).
func (r LoginUserRequest) Identifier() string {
	if r.Login != "" {
		return r.Login
	}
	return r.Email
}

type LoginUserResponse struct {
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
	User         UserResponse `json:"user"`
}

// OAuthTokenResponse is the OAuth2 password-grant response used by Swagger Authorize.
type OAuthTokenResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType    string `json:"token_type" example:"bearer"`
	ExpiresIn    int    `json:"expires_in" example:"900"`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type UserResponse struct {
	ID    int             `json:"id"`
	Name  string          `json:"name"`
	Login string          `json:"login"`
	Email string          `json:"email"`
	Role  schema.UserRole `json:"role"`
}

type NewUserRequest struct {
	Name     string `json:"name" binding:"required" example:"John Doe"`
	Login    string `json:"login" binding:"required" example:"johndoe"`
	Email    string `json:"email" binding:"required" example:"john.doe@example.com"`
	Password string `json:"password" binding:"required" example:"password"`
}
