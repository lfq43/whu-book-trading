package middleware

import (
	"net/http"
	"strings"

	"book-trading/backend/internal/models"
	"book-trading/backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware JWT 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 从请求头获取 Authorization
		// 格式: Authorization: Bearer <token>
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.Response{
				Code:    401,
				Message: "未提供认证令牌",
				Data:    nil,
			})
			c.Abort() // 中止后续处理
			return
		}

		// 2. 解析 token（去掉 "Bearer " 前缀）
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, models.Response{
				Code:    401,
				Message: "认证令牌格式错误",
				Data:    nil,
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 3. 验证 token
		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.Response{
				Code:    401,
				Message: "无效的认证令牌: " + err.Error(),
				Data:    nil,
			})
			c.Abort()
			return
		}

		// 4. 将用户信息存入上下文，后续的控制器可以获取
		c.Set("userID", claims.UserID) //后续使用的userid
		c.Set("username", claims.Username)

		// 5. 继续处理请求
		c.Next()
	}
}
