package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	secretKey     string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
	issuer        string
}

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func NewJWTManager(secretKey string, accessExpiry, refreshExpiry time.Duration, issuer string) *JWTManager {
	return &JWTManager{
		secretKey:     secretKey,
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
		issuer:        issuer,
	}
}

func (m *JWTManager) GenerateTokenPair(userID, username, email, role string) (*TokenPair, error) {
	accessToken, accessExp, err := m.generateToken(userID, username, email, role, m.accessExpiry)
	if err != nil {
		return nil, err
	}

	refreshToken, _, err := m.generateToken(userID, username, email, role, m.refreshExpiry)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    accessExp.Unix(),
	}, nil
}

func (m *JWTManager) generateToken(userID, username, email, role string, expiry time.Duration) (string, *time.Time, error) {
	expirationTime := time.Now().Add(expiry)

	claims := &Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    m.issuer,
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.secretKey))
	if err != nil {
		return "", nil, err
	}

	return tokenString, &expirationTime, nil
}

func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (m *JWTManager) RefreshToken(refreshToken string) (*TokenPair, error) {
	claims, err := m.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	if claims.ExpiresAt == nil || claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("refresh token expired")
	}

	return m.GenerateTokenPair(claims.UserID, claims.Username, claims.Email, claims.Role)
}

func (m *JWTManager) ExtractClaims(tokenString string) (*Claims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok {
		return claims, nil
	}

	return nil, errors.New("failed to extract claims")
}

func (m *JWTManager) GetExpirationTime(tokenString string) (*time.Time, error) {
	claims, err := m.ExtractClaims(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.ExpiresAt != nil {
		return &claims.ExpiresAt.Time, nil
	}

	return nil, errors.New("no expiration time in token")
}
