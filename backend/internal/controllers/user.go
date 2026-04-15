package controllers

import (
	"context"
	"net/http"
	"strconv"

	"book-trading/backend/internal/config"
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

	// 如果填写了邮箱，则必须提供验证码并校验（5分钟有效）
	if req.Email != "" {
		if req.VerificationCode == "" {
			c.JSON(http.StatusBadRequest, models.Response{
				Code:    400,
				Message: "请提供邮箱验证码",
				Data:    nil,
			})
			return
		}
		ctx := context.Background()
		key := "verify:" + req.Email
		storedCode, err := database.RedisClient.Get(ctx, key).Result()
		if err != nil {
			c.JSON(http.StatusBadRequest, models.Response{
				Code:    400,
				Message: "验证码无效或已过期",
				Data:    nil,
			})
			return
		}
		if storedCode != req.VerificationCode {
			c.JSON(http.StatusBadRequest, models.Response{
				Code:    400,
				Message: "验证码错误",
				Data:    nil,
			})
			return
		}
		// 验证通过后删除验证码
		_ = database.RedisClient.Del(ctx, key)
	}

	// 2. 检查账号名是否已被使用
	var existingUser models.User
	result := database.DB.Where("account = ?", req.Account).First(&existingUser)
	if result.Error == nil {
		c.JSON(http.StatusConflict, models.Response{
			Code:    409,
			Message: "账号名已被注册",
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
		Account:  req.Account,
		Username: req.Username,
		Password: hashedPassword,
		Email:    req.Email,
		Avatar:   "/public/defaultavatar.jpg",
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
			"account":  user.Account,
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

	// 2. 检查是否是配置里的管理员账号
	if req.Account == config.AppConfig.AdminAccount && req.Password == config.AppConfig.AdminPassword {
		adminUser := models.User{
			ID:       4,
			Account:  config.AppConfig.AdminAccount,
			Username: "管理员",
			IsAdmin:  true,
		}

		token, err := utils.GenerateToken(4, "管理员", true)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Code:    500,
				Message: "服务器错误：生成令牌失败",
				Data:    nil,
			})
			return
		}

		c.JSON(http.StatusOK, models.Response{
			Code:    0,
			Message: "登录成功",
			Data: models.LoginResponse{
				Token: token,
				User:  adminUser,
			},
		})
		return
	}

	// 3. 查找用户
	var user models.User
	result := database.DB.Where("account = ?", req.Account).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, models.Response{
			Code:    401,
			Message: "账号或密码错误",
			Data:    nil,
		})
		return
	}

	// 4. 禁用用户不能登录
	if user.IsBanned {
		c.JSON(http.StatusForbidden, models.Response{
			Code:    403,
			Message: "账号已被封禁，请联系管理员",
			Data:    nil,
		})
		return
	}

	// 5. 验证密码
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, models.Response{
			Code:    401,
			Message: "账号或密码错误",
			Data:    nil,
		})
		return
	}

	// 6. 生成 JWT Token
	token, err := utils.GenerateToken(user.ID, user.Username, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "服务器错误：生成令牌失败",
			Data:    nil,
		})
		return
	}

	// 7. 返回成功和 token
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

// GetUserProfile 获取指定用户的公开信息（用于个人空间）
func GetUserProfile(c *gin.Context) {
	userIDStr := c.Param("id")
	var userID uint
	if _, ok := c.Params.Get("id"); ok {
		// 解析用户ID
		userID = parseUint(userIDStr)
	}

	if userID == 0 {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "无效的用户ID",
			Data:    nil,
		})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "用户不存在",
			Data:    nil,
		})
		return
	}

	// 获取用户发布的批次
	var batches []models.Batch
	database.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&batches)

	c.JSON(http.StatusOK, models.Response{
		Code:    0,
		Message: "success",
		Data: gin.H{
			"user":    user,
			"batches": batches,
		},
	})
}

// GetAllUsers 管理员获取全部用户
func GetAllUsers(c *gin.Context) {
	var users []models.User
	database.DB.Order("created_at DESC").Find(&users)

	c.JSON(http.StatusOK, models.Response{
		Code:    0,
		Message: "success",
		Data:    users,
	})
}

// BanUser 管理员封禁用户
func BanUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "无效的用户ID",
			Data:    nil,
		})
		return
	}

	var user models.User
	if err := database.DB.First(&user, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "用户不存在",
			Data:    nil,
		})
		return
	}

	if user.Account == config.AppConfig.AdminAccount {
		c.JSON(http.StatusForbidden, models.Response{
			Code:    403,
			Message: "不能封禁管理员账号",
			Data:    nil,
		})
		return
	}

	user.IsBanned = true
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "封禁失败",
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    0,
		Message: "封禁成功",
		Data:    nil,
	})
}

// UpdateProfile 更新用户信息
func UpdateProfile(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"omitempty,min=2,max=50"`
		Avatar   string `json:"avatar" binding:"omitempty"`
		Email    string `json:"email" binding:"omitempty,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "参数错误: " + err.Error(),
			Data:    nil,
		})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.Response{
			Code:    401,
			Message: "未认证",
			Data:    nil,
		})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "用户不存在",
			Data:    nil,
		})
		return
	}

	// 如果要修改用户名，检查是否已被使用
	if req.Username != "" && req.Username != user.Username {
		var existingUser models.User
		if err := database.DB.Where("username = ? AND id != ?", req.Username, userID).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, models.Response{
				Code:    409,
				Message: "用户名已被占用",
				Data:    nil,
			})
			return
		}
		user.Username = req.Username
	}

	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}

	if req.Email != "" {
		user.Email = req.Email
	}

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "更新失败: " + err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    0,
		Message: "更新成功",
		Data:    user,
	})
}

// 辅助函数：解析 uint
func parseUint(s string) uint {
	var result uint
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		result = result*10 + uint(c-'0')
	}
	return result
}
