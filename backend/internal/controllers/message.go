package controllers

import (
	"net/http"
	"strconv"
	"time"

	"book-trading/backend/internal/database"
	"book-trading/backend/internal/models"

	"github.com/gin-gonic/gin"
)

// SendMessage 发送消息
func SendMessage(c *gin.Context) {
	var req struct {
		ToUserID uint   `json:"to_user_id" binding:"required"`
		Content  string `json:"content" binding:"required,min=1,max=1000"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "参数错误: " + err.Error(),
			Data:    nil,
		})
		return
	}

	// 获取当前登录用户ID
	fromUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.Response{
			Code:    401,
			Message: "请先登录",
			Data:    nil,
		})
		return
	}

	// 不能给自己发消息
	if fromUserID.(uint) == req.ToUserID {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "不能给自己发送消息",
			Data:    nil,
		})
		return
	}

	// 检查接收者是否存在
	var toUser models.User
	if err := database.DB.First(&toUser, req.ToUserID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "接收者不存在",
			Data:    nil,
		})
		return
	}

	// 创建消息
	message := models.Message{
		FromUserID: fromUserID.(uint),
		ToUserID:   req.ToUserID,
		Content:    req.Content,
		IsRead:     false,
	}

	if err := database.DB.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "发送失败: " + err.Error(),
			Data:    nil,
		})
		return
	}

	// 预加载发送者信息
	database.DB.Preload("FromUser").First(&message, message.ID)

	c.JSON(http.StatusOK, models.Response{
		Code:    0,
		Message: "发送成功",
		Data:    message,
	})
}

// GetConversation 获取与某个用户的聊天记录
func GetConversation(c *gin.Context) {
	// 获取对方用户ID
	otherUserIDStr := c.Param("userId")
	otherUserID, err := strconv.ParseUint(otherUserIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "无效的用户ID",
			Data:    nil,
		})
		return
	}

	// 获取当前用户ID
	currentUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.Response{
			Code:    401,
			Message: "请先登录",
			Data:    nil,
		})
		return
	}

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	// 查询聊天记录（双方之间的所有消息）
	var messages []models.Message
	offset := (page - 1) * pageSize

	query := database.DB.Where(
		"(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)",
		currentUserID, otherUserID,
		otherUserID, currentUserID,
	).Order("created_at DESC")

	// 统计总数
	var total int64
	query.Count(&total)

	// 分页查询
	query.Offset(offset).Limit(pageSize).Preload("FromUser").Find(&messages)

	// 将消息按时间正序返回（方便前端显示）
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	// 标记所有收到的消息为已读
	database.DB.Model(&models.Message{}).
		Where("to_user_id = ? AND from_user_id = ? AND is_read = ?", currentUserID, otherUserID, false).
		Update("is_read", true)

	c.JSON(http.StatusOK, models.Response{
		Code:    0,
		Message: "success",
		Data: gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
			"messages":  messages,
		},
	})
}

// GetUnreadCount 获取未读消息总数
func GetUnreadCount(c *gin.Context) {
	currentUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.Response{
			Code:    401,
			Message: "请先登录",
			Data:    nil,
		})
		return
	}

	var count int64
	database.DB.Model(&models.Message{}).
		Where("to_user_id = ? AND is_read = ?", currentUserID, false).
		Count(&count)

	c.JSON(http.StatusOK, models.Response{
		Code:    0,
		Message: "success",
		Data:    count,
	})
}

// GetConversationList 获取对话列表（所有聊过天的人）
func GetConversationList(c *gin.Context) {
	currentUserID, _ := c.Get("userID")

	// 1. 获取所有对话过的用户ID及最后消息时间
	var conversations []struct {
		UserID      uint
		LastMessage string
		LastTime    time.Time
	}

	// 使用 GORM 的子查询
	subQuery := database.DB.Model(&models.Message{}).
		Select("CASE WHEN from_user_id = ? THEN to_user_id ELSE from_user_id END as user_id, MAX(created_at) as last_time", currentUserID).
		Where("from_user_id = ? OR to_user_id = ?", currentUserID, currentUserID).
		Group("user_id")

	database.DB.Table("(?) as latest", subQuery).
		Select("latest.user_id, m.content as last_message, latest.last_time").
		Joins("LEFT JOIN messages m ON (m.from_user_id = ? AND m.to_user_id = latest.user_id AND m.created_at = latest.last_time) OR (m.from_user_id = latest.user_id AND m.to_user_id = ? AND m.created_at = latest.last_time)", currentUserID, currentUserID).
		Scan(&conversations)

	// 2. 组装返回数据（循环添加用户信息和未读数）
	var result []gin.H
	for _, conv := range conversations {
		var user models.User
		database.DB.First(&user, conv.UserID)

		var unreadCount int64
		database.DB.Model(&models.Message{}).
			Where("from_user_id = ? AND to_user_id = ? AND is_read = ?", conv.UserID, currentUserID, false).
			Count(&unreadCount)

		result = append(result, gin.H{
			"user_id":      conv.UserID,
			"username":     user.Username,
			"avatar":       user.Avatar,
			"last_message": conv.LastMessage,
			"last_time":    conv.LastTime,
			"unread_count": unreadCount,
		})
	}

	c.JSON(200, result)
}
