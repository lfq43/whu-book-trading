package models

import (
	"time"
)

// Batch 发布批次（一批书）
type Batch struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Title       string `gorm:"size:200;not null;index" json:"title"`
	Description string `gorm:"type:text" json:"description"`
	Image       string `gorm:"type:json" json:"image"`

	// 未售出的书籍列表（JSON 字符串数组）
	BookNames string `gorm:"type:json" json:"book_names"` // ["高等数学", "大学英语"]

	// 已售出的书籍列表（JSON 字符串数组）
	SoldBookNames string `gorm:"type:json" json:"sold_book_names"` // ["政治", "专业课"]

	// 联系方式
	Contact string `gorm:"size:100" json:"contact"`

	// 状态（根据 book_names 长度自动计算）
	Status string `gorm:"size:20;default:'available';index" json:"status"` // available/partial/sold

	UserID uint `gorm:"not null;index" json:"user_id"`
	User   User `gorm:"foreignKey:UserID" json:"user,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Batch) TableName() string {
	return "batches"
}
