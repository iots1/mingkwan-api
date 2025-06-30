package adapters

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims defines the JWT claims structure.
type Claims struct {
	UserID string `json:"userId"`
	jwt.RegisteredClaims
}

// JWTTokenGenerator defines the interface for generating and parsing JWTs.
type JWTTokenGenerator interface {
	GenerateTokens(userID string) (accessToken, refreshToken string, err error)
	ParseAccessToken(tokenString string) (*Claims, error)
	ParseRefreshToken(tokenString string) (*Claims, error)
}

// JWTTokenConfig holds configuration for JWT generation.
type JWTTokenConfig struct {
	Secret           string
	AccessExpMinutes int
	RefreshExpDays   int
}

// JWTGenerator implements JWTTokenGenerator.
type JWTGenerator struct {
	config JWTTokenConfig
}

func NewJWTTokenGenerator(secret string) JWTTokenGenerator {
	return &JWTGenerator{
		config: JWTTokenConfig{
			Secret:           secret,
			AccessExpMinutes: 15, // e.g., 15 minutes
			RefreshExpDays:   7,  // e.g., 7 days
		},
	}
}

func (j *JWTGenerator) GenerateTokens(userID string) (accessToken, refreshToken string, err error) {
	// Access Token
	accessClaims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * time.Duration(j.config.AccessExpMinutes))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(j.config.Secret))
	if err != nil {
		return "", "", err
	}

	// Refresh Token
	refreshClaims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * time.Duration(j.config.RefreshExpDays))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(j.config.Secret))
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (j *JWTGenerator) ParseAccessToken(tokenString string) (*Claims, error) {
	// ... Parse Access Token logic
	return nil, nil // Placeholder
}

func (j *JWTGenerator) ParseRefreshToken(tokenString string) (*Claims, error) {
	// ... Parse Refresh Token logic
	return nil, nil // Placeholder
}
