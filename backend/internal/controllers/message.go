package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"book-trading/backend/internal/database"
	"book-trading/backend/internal/models"
	"book-trading/backend/internal/sse"
	"book-trading/backend/internal/utils"
	"book-trading/backend/internal/ws"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func createMessage(fromUserID, toUserID uint, content string) (*models.Message, error) {
	if fromUserID == toUserID {
		return nil, errors.New("不能给自己发送消息")
	}

	var toUser models.User
	if err := database.DB.First(&toUser, toUserID).Error; err != nil {
		return nil, err
	}

	message := models.Message{
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Content:    content,
		IsRead:     false,
	}

	if err := database.DB.Create(&message).Error; err != nil {
		return nil, err
	}

	updateConversation(fromUserID, toUserID, content, false)
	updateConversation(toUserID, fromUserID, content, true)

	database.DB.Preload("FromUser").First(&message, message.ID)
	notifyUnreadCount(toUserID)
	return &message, nil
}

func getTotalUnreadCount(userID uint) int64 {
	var totalUnread int64
	database.DB.Model(&models.Conversation{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(unread_count), 0)").
		Scan(&totalUnread)
	return totalUnread
}

func notifyUnreadCount(userID uint) {
	totalUnread := getTotalUnreadCount(userID)
	sse.DefaultManager.Send(userID, int(totalUnread))
}

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

	fromUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.Response{
			Code:    401,
			Message: "请先登录",
			Data:    nil,
		})
		return
	}

	message, err := createMessage(fromUserID.(uint), req.ToUserID, req.Content)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.Response{
				Code:    404,
				Message: "接收者不存在",
				Data:    nil,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "发送失败: " + err.Error(),
			Data:    nil,
		})
		return
	}

	ws.DefaultHub.SendToUser(req.ToUserID, gin.H{"type": "message_received", "data": message})

	c.JSON(http.StatusOK, models.Response{
		Code:    0,
		Message: "发送成功",
		Data:    message,
	})
}

// updateConversation 更新对话记录
func updateConversation(userID, otherUserID uint, lastMessage string, incrementUnread bool) {
	var conv models.Conversation

	// 查找是否存在对话记录
	result := database.DB.Where("user_id = ? AND other_user_id = ?", userID, otherUserID).First(&conv)

	if result.Error != nil {
		// 不存在，创建新记录
		conv = models.Conversation{
			UserID:      userID,
			OtherUserID: otherUserID,
			LastMessage: lastMessage,
			LastTime:    database.DB.NowFunc(),
			UnreadCount: 0,
		}
		if incrementUnread {
			conv.UnreadCount = 1
		}
		database.DB.Create(&conv)
	} else {
		// 存在，更新记录
		updates := map[string]interface{}{
			"last_message": lastMessage,
			"last_time":    database.DB.NowFunc(),
		}
		if incrementUnread {
			updates["unread_count"] = conv.UnreadCount + 1
		} else {
			updates["unread_count"] = 0 // 发送者视角，未读数为0
		}
		database.DB.Model(&conv).Updates(updates)
	}
}

// GetConversation 获取与某个用户的聊天记录（滑动加载）
func GetConversation(c *gin.Context) {
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

	currentUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.Response{
			Code:    401,
			Message: "请先登录",
			Data:    nil,
		})
		return
	}

	// 游标分页参数
	// before_id: 查询比这个ID更旧的消息（用于加载更多）
	// limit: 每页数量
	beforeIDStr := c.Query("before_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if limit < 1 {
		limit = 20
	}
	if limit > 50 {
		limit = 50 // 限制最大50条
	}

	// 构建查询
	query := database.DB.Where(
		"(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)",
		currentUserID, otherUserID,
		otherUserID, currentUserID,
	)

	// 如果提供了 before_id，查询更旧的消息
	if beforeIDStr != "" {
		beforeID, err := strconv.ParseUint(beforeIDStr, 10, 32)
		if err == nil {
			query = query.Where("id < ?", beforeID)
		}
	}

	// 查询消息（按 ID 倒序，最新的在前面）
	var messages []models.Message
	err = query.Order("id DESC").
		Limit(limit).
		Preload("FromUser").
		Find(&messages).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询失败: " + err.Error(),
			Data:    nil,
		})
		return
	}

	// 判断是否还有更多
	hasMore := len(messages) == limit

	// 将消息按时间正序返回（方便前端显示，最早的在前）
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	// 获取第一条消息的ID（用于下次加载更多）
	var nextBeforeID uint = 0
	if len(messages) > 0 {
		nextBeforeID = messages[0].ID
	}

	// 将该对话中所有未读消息标记为已读
	database.DB.Model(&models.Message{}).
		Where("to_user_id = ? AND from_user_id = ? AND is_read = ?", currentUserID, otherUserID, false).
		Update("is_read", true)

	// 重置该对话的未读计数
	database.DB.Model(&models.Conversation{}).
		Where("user_id = ? AND other_user_id = ?", currentUserID, otherUserID).
		Update("unread_count", 0)

	notifyUnreadCount(currentUserID.(uint))

	c.JSON(http.StatusOK, models.Response{
		Code:    0,
		Message: "success",
		Data: gin.H{
			"messages":       messages,
			"has_more":       hasMore,
			"next_before_id": nextBeforeID,
		},
	})
}

// GetConversationList 获取对话列表（简单查询）
func GetConversationList(c *gin.Context) {
	currentUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.Response{
			Code:    401,
			Message: "请先登录",
			Data:    nil,
		})
		return
	}

	var conversations []models.Conversation

	// 简单查询：获取当前用户的所有对话，按最后消息时间倒序
	err := database.DB.
		Where("user_id = ?", currentUserID).
		Preload("OtherUser").
		Order("last_time DESC").
		Find(&conversations).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询失败",
			Data:    nil,
		})
		return
	}

	// 转换格式，方便前端使用
	result := make([]gin.H, 0, len(conversations))
	for _, conv := range conversations {
		result = append(result, gin.H{
			"user_id":      conv.OtherUserID,
			"username":     conv.OtherUser.Username,
			"avatar":       conv.OtherUser.Avatar,
			"last_message": conv.LastMessage,
			"last_time":    conv.LastTime,
			"unread_count": conv.UnreadCount,
		})
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    0,
		Message: "success",
		Data:    result,
	})
}

// GetUnreadCount 获取总未读消息数
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

	var totalUnread int64
	database.DB.Model(&models.Conversation{}).
		Where("user_id = ?", currentUserID).
		Select("COALESCE(SUM(unread_count), 0)").
		Scan(&totalUnread)

	c.JSON(http.StatusOK, models.Response{
		Code:    0,
		Message: "success",
		Data:    totalUnread,
	})
}

func SubscribeUnreadSSE(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if token == "" {
		c.JSON(http.StatusUnauthorized, models.Response{
			Code:    401,
			Message: "未提供认证令牌",
			Data:    nil,
		})
		return
	}

	claims, err := utils.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.Response{
			Code:    401,
			Message: "无效的认证令牌: " + err.Error(),
			Data:    nil,
		})
		return
	}

	userID := claims.UserID

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "当前环境不支持 SSE",
			Data:    nil,
		})
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)

	ch := sse.DefaultManager.Register(userID)
	defer sse.DefaultManager.Unregister(userID, ch)

	// 首次发送当前未读数量
	initialCount := getTotalUnreadCount(userID)
	fmt.Fprintf(c.Writer, "event: unread\ndata: %d\n\n", initialCount)
	flusher.Flush()

	ctx := c.Request.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case count, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(c.Writer, "event: unread\ndata: %d\n\n", count)
			flusher.Flush()
		}
	}
}
