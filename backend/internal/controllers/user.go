package controllers

import (
	"net/http"

	"book-trading/backend/internal/database"
	"book-trading/backend/internal/models"
	"book-trading/backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// Register 用户注册
func Register(c *gin.Context) {
	var req models.RegisterRequest

	// 1. 解析并验证请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "参数错误: " + err.Error(),
			Data:    nil,
		})
		return
	}

	// 2. 检查用户名是否已被使用
	var existingUser models.User
	result := database.DB.Where("username = ?", req.Username).First(&existingUser)
	if result.Error == nil {
		// 用户已存在
		c.JSON(http.StatusConflict, models.Response{
			Code:    409,
			Message: "用户名已被注册",
			Data:    nil,
		})
		return
	}

	// 3. 加密密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "服务器错误：密码加密失败",
			Data:    nil,
		})
		return
	}

	// 4. 创建用户对象
	user := models.User{
		Username: req.Username,
		Password: hashedPassword,
		Email:    req.Email,
	}

	// 5. 保存到数据库
	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "服务器错误：创建用户失败",
			Data:    nil,
		})
		return
	}

	// 6. 返回成功（不返回密码）
	c.JSON(http.StatusOK, models.Response{
		Code:    0,
		Message: "注册成功",
		Data: gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

// Login 用户登录
func Login(c *gin.Context) {
	var req models.LoginRequest

	// 1. 解析请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "参数错误: " + err.Error(),
			Data:    nil,
		})
		return
	}

	// 2. 查找用户
	var user models.User
	result := database.DB.Where("username = ?", req.Username).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, models.Response{
			Code:    401,
			Message: "用户名或密码错误",
			Data:    nil,
		})
		return
	}

	// 3. 验证密码
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, models.Response{
			Code:    401,
			Message: "用户名或密码错误",
			Data:    nil,
		})
		return
	}

	// 4. 生成 JWT Token
	token, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "服务器错误：生成令牌失败",
			Data:    nil,
		})
		return
	}

	// 5. 返回成功和 token
	c.JSON(http.StatusOK, models.Response{
		Code:    0,
		Message: "登录成功",
		Data: models.LoginResponse{
			Token: token,
			User:  user,
		},
	})
}

// GetProfile 获取当前登录用户的信息（需要认证）
func GetProfile(c *gin.Context) {
	// 从上下文中获取用户信息（JWT中间件会设置）
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.Response{
			Code:    401,
			Message: "未认证",
			Data:    nil,
		})
		return
	}

	// 查询用户信息
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "用户不存在",
			Data:    nil,
		})
		return
	}

	// 返回用户信息
	c.JSON(http.StatusOK, models.Response{
		Code:    0,
		Message: "success",
		Data:    user,
	})
}
