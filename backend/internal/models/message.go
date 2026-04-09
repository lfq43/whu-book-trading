package models

import (
	"time"
)

// Message 聊天消息
type Message struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	FromUserID uint      `gorm:"not null;index" json:"from_user_id"` // 发送者
	ToUserID   uint      `gorm:"not null;index" json:"to_user_id"`   // 接收者
	Content    string    `gorm:"type:text;not null" json:"content"`  // 消息内容
	IsRead     bool      `gorm:"default:false" json:"is_read"`       // 是否已读
	CreatedAt  time.Time `json:"created_at"`

	// 关联（不存数据库，用于查询时填充）
	FromUser User `gorm:"foreignKey:FromUserID" json:"from_user,omitempty"`
	ToUser   User `gorm:"foreignKey:ToUserID" json:"to_user,omitempty"`
}

func (Message) TableName() string {
	return "messages"
}
