package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// TokenType Token类型
type TokenType int

const (
	AccessToken  TokenType = iota // 访问令牌
	RefreshToken                  // 刷新令牌
)

// CustomClaims 自定义JWT声明
type CustomClaims struct {
	UserId    int64     `json:"userId"`
	Username  string    `json:"username"`
	TokenType TokenType `json:"tokenType"`
	jwt.RegisteredClaims
}

// TokenPair Token对
type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

// GenerateToken 生成JWT Token
func GenerateToken(userId int64, username string, secret string, expireSeconds int64, tokenType TokenType) (string, error) {
	claims := CustomClaims{
		UserId:    userId,
		Username:  username,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireSeconds) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "im-auth",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken 解析JWT Token
func ParseToken(tokenString string, secret string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token已过期")
		}
		return nil, errors.New("token无效")
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("token无效")
}

// GenerateTokenPair 生成Token对
func GenerateTokenPair(userId int64, username string, accessSecret string, accessExpire int64, refreshSecret string, refreshExpire int64) (*TokenPair, error) {
	// 生成访问令牌
	accessToken, err := GenerateToken(userId, username, accessSecret, accessExpire, AccessToken)
	if err != nil {
		return nil, err
	}

	// 生成刷新令牌
	refreshToken, err := GenerateToken(userId, username, refreshSecret, refreshExpire, RefreshToken)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    accessExpire,
	}, nil
}

// ValidateTokenType 验证Token类型
func ValidateTokenType(claims *CustomClaims, expectedType TokenType) bool {
	return claims.TokenType == expectedType
}
