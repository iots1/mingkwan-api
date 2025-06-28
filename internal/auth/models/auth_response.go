package models

import userModel "github.com/iots1/mingkwan-api/internal/user/models"

type AuthResponse struct {
	User         *userModel.UserResponse `json:"user,omitempty"`
	AccessToken  string                  `json:"access_token"`
	RefreshToken string                  `json:"refresh_token,omitempty"`
	ExpiresIn    int64                   `json:"expires_in"`
	TokenType    string                  `json:"token_type"`
}
