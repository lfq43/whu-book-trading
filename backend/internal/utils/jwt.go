package utils

import (
	"errors"
	"time"

	"book-trading/backend/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 定义 JWT 中包含的信息
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 JWT 令牌
func GenerateToken(userID uint, username string) (string, error) {
	// 设置令牌的有效期为 24 小时
	expirationTime := time.Now().Add(24 * time.Hour)

	// 创建 Claims（包含用户信息和过期时间）
	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// 使用 HS256 算法签名，密钥从配置中获取
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名并返回 token 字符串
	return token.SignedString([]byte(config.AppConfig.JWTSecret))
}

// ParseToken 解析并验证 JWT 令牌
func ParseToken(tokenString string) (*Claims, error) {
	// 解析 token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法是否正确
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(config.AppConfig.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	// 验证 token 是否有效
	if claims, ok := token.Claims.(*Claims); ok && token.Valid { //检查token.Claims是否是自定义的Claims类型
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
