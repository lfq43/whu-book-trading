package models

import (
	"time"
)

// Batch 发布批次（一批书）
type Batch struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Title       string `gorm:"size:200;not null;index" json:"title"`
	Description string `gorm:"type:text" json:"description"`
	Image       string `gorm:"size:500" json:"image"`

	// 书籍名字列表（存 JSON 字符串）
	// 例如：["高等数学", "大学英语", "政治"]
	BookNames string `gorm:"type:json" json:"book_names"`

	// 售出状态列表（存 JSON 字符串）
	// 例如：[false, false, true]
	SoldStatus string `gorm:"type:json" json:"sold_status"`

	// 联系方式
	Contact string `gorm:"size:100" json:"contact"`

	// 状态
	Status string `gorm:"size:20;default:'available';index" json:"status"`

	UserID uint `gorm:"not null;index" json:"user_id"`
	User   User `gorm:"foreignKey:UserID" json:"user,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Batch) TableName() string {
	return "batches"
}
