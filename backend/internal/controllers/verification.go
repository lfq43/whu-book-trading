package controllers

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"book-trading/backend/internal/database"
	"book-trading/backend/internal/models"
	"book-trading/backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type sendCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// SendVerificationCode 发送邮箱验证码（5分钟有效）
func SendVerificationCode(c *gin.Context) {
	var req sendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "参数错误: " + err.Error(),
			Data:    nil,
		})
		return
	}

	// 生成 6 位数字验证码（使用 crypto/rand）
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "生成验证码失败",
			Data:    nil,
		})
		return
	}
	code := fmt.Sprintf("%06d", n.Int64())

	// 存入 Redis，5分钟后过期
	key := "verify:" + req.Email
	ctx := context.Background()
	if err := database.RedisClient.Set(ctx, key, code, 5*time.Minute).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "保存验证码失败",
			Data:    nil,
		})
		return
	}

	// 发送邮件
	subject := "书籍交易平台 — 注册验证码"
	body := fmt.Sprintf("<p>您用于注册的验证码为：<strong>%s</strong></p><p>5分钟内有效，请勿泄露。</p>", code)
	if err := utils.SendMail(req.Email, subject, body); err != nil {
		// 发送失败则删除 Redis 中的验证码
		_ = database.RedisClient.Del(ctx, key)
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "发送邮件失败: " + err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    0,
		Message: "验证码已发送，请查收",
		Data:    nil,
	})
}
