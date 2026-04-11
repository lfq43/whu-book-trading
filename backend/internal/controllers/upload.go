package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"book-trading/backend/internal/database"
	"book-trading/backend/internal/models"

	"github.com/gin-gonic/gin"
)

// UploadImage 上传图片
func UploadImage(c *gin.Context) {
	// 获取当前用户
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.Response{
			Code:    401,
			Message: "请先登录",
			Data:    nil,
		})
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "请选择图片文件",
			Data:    nil,
		})
		return
	}

	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "只支持 JPG、PNG、GIF、WEBP 格式的图片",
			Data:    nil,
		})
		return
	}

	// 检查文件大小（限制 5MB）
	if file.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "图片大小不能超过 5MB",
			Data:    nil,
		})
		return
	}

	// 生成唯一的文件名
	filename := fmt.Sprintf("%d_%d%s", userID.(uint), time.Now().UnixNano(), ext)

	// 保存路径（相对于项目根目录）
	// 创建上传目录
	uploadDir := "uploads"

	// 保存文件
	savePath := filepath.Join(uploadDir, filename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "保存图片失败: " + err.Error(),
			Data:    nil,
		})
		return
	}

	// 返回图片URL
	imageURL := "/uploads/" + filename

	c.JSON(http.StatusOK, models.Response{
		Code:    0,
		Message: "上传成功",
		Data: gin.H{
			"url": imageURL,
		},
	})
}

// UploadAvatar 上传头像
func UploadAvatar(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.Response{
			Code:    401,
			Message: "请先登录",
			Data:    nil,
		})
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "请选择图片文件",
			Data:    nil,
		})
		return
	}

	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "只支持 JPG、PNG、GIF、WEBP 格式的图片",
			Data:    nil,
		})
		return
	}

	// 检查文件大小（限制 2MB，头像小一点）
	if file.Size > 2*1024*1024 {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "头像大小不能超过 2MB",
			Data:    nil,
		})
		return
	}

	// 生成唯一的文件名
	filename := fmt.Sprintf("avatar_%d_%d%s", userID.(uint), time.Now().UnixNano(), ext)

	// 保存文件
	uploadDir := "uploads/avatars"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "创建目录失败",
			Data:    nil,
		})
		return
	}

	savePath := filepath.Join(uploadDir, filename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "保存头像失败: " + err.Error(),
			Data:    nil,
		})
		return
	}

	// 返回图片URL
	imageURL := "/uploads/avatars/" + filename

	// 更新用户头像
	var user models.User
	if err := database.DB.First(&user, userID).Error; err == nil {
		user.Avatar = imageURL
		database.DB.Save(&user)
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    0,
		Message: "上传成功",
		Data: gin.H{
			"url": imageURL,
		},
	})
}
