package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Service struct {
	jwtSecret  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func NewService(jwtSecret string, accessTTLMinutes, refreshTTLDays int) *Service {
	return &Service{
		jwtSecret:  []byte(jwtSecret),
		accessTTL:  time.Duration(accessTTLMinutes) * time.Minute,
		refreshTTL: time.Duration(refreshTTLDays) * 24 * time.Hour,
	}
}

func (s *Service) GenerateTokens(userID uuid.UUID, username string) (*TokenPair, error) {
	now := time.Now()

	accessClaims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "sync",
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenStr, err := accessToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	refreshClaims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshTTL)),
		IssuedAt:  jwt.NewNumericDate(now),
		Issuer:    "sync",
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenStr, err := refreshToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenStr,
		RefreshToken: refreshTokenStr,
		ExpiresIn:    int(s.accessTTL.Seconds()),
	}, nil
}

func (s *Service) ValidateAccessToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
