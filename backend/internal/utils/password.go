package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword 把明文密码加密成哈希值
// 例如：输入 "123456" -> 输出 "$2a$10$N9qo8uLOickgx2ZMRZoMy.Mr/.cZqV5KvUqN5KvUqN5KvUqN5KvU"
func HashPassword(password string) (string, error) {
	// bcrypt.DefaultCost 是默认的加密强度，数值越大越安全但越慢
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash 验证密码是否匹配
// 输入：明文密码 + 数据库里的哈希值 -> 返回是否匹配
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
