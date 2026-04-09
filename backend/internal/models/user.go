package models

import (
	"time"
)

// User 用户模型 - 对应数据库中的 users 表
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`                         // 主键ID
	Username  string    `gorm:"uniqueIndex;size:50;not null" json:"username"` // 用户名，唯一，最大50字符，不能为空
	Password  string    `gorm:"size:255;not null" json:"-"`                   // 密码，json:"-" 表示返回JSON时隐藏密码
	Email     string    `gorm:"size:100" json:"email"`                        // 邮箱
	Avatar    string    `gorm:"size:255" json:"avatar"`                       // 头像URL
	CreatedAt time.Time `json:"created_at"`                                   // 创建时间
	UpdatedAt time.Time `json:"updated_at"`                                   // 更新时间
}

// TableName 指定数据库表名（GORM默认会使用复数形式，这里明确指定）
func (User) TableName() string {
	return "users"
}
