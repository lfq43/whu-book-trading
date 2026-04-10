package models

import (
	"time"
)

// Conversation 对话列表（冗余表，简化查询）
type Conversation struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;index:idx_user_pair,unique" json:"user_id"`       // 当前用户ID
	OtherUserID uint      `gorm:"not null;index:idx_user_pair,unique" json:"other_user_id"` // 对方用户ID
	LastMessage string    `gorm:"type:text" json:"last_message"`                            // 最后一条消息内容
	LastTime    time.Time `json:"last_time"`                                                // 最后消息时间
	UnreadCount int       `gorm:"default:0" json:"unread_count"`                            // 未读消息数量（针对 user_id）

	// 关联（不存数据库）
	OtherUser User `gorm:"foreignKey:OtherUserID" json:"other_user,omitempty"`
}

func (Conversation) TableName() string {
	return "conversations"
}
