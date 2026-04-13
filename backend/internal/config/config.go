package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	RedisHost string
	RedisPort string
	RedisPwd  string

	JWTSecret string

	// SMTP 配置（用于发送邮箱验证码）
	SMTPHost string
	SMTPPort string
	SMTPUser string
	SMTPPass string
	SMTPFrom string

	AdminAccount  string
	AdminPassword string

	ServerPort string
}

var AppConfig *Config

func LoadConfig() {
	err := godotenv.Load(".env") //load env file
	if err != nil {
		log.Println("No .env file found, using system env")
	}

	AppConfig = &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "bookuser"),
		DBPassword: getEnv("DB_PASSWORD", "book123"),
		DBName:     getEnv("DB_NAME", "book_trading"),

		RedisHost: getEnv("REDIS_HOST", "localhost"),
		RedisPort: getEnv("REDIS_PORT", "6379"),
		RedisPwd:  getEnv("REDIS_PASSWORD", ""),

		JWTSecret: getEnv("JWT_SECRET", "default-secret-key"),

		AdminAccount:  getEnv("ADMIN_ACCOUNT", "lfq43"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin123"),

		ServerPort: getEnv("SERVER_PORT", "8082"),
		// SMTP 默认不填，建议在 .env 中配置：SMTP_HOST、SMTP_PORT、SMTP_USER、SMTP_PASS、SMTP_FROM
		SMTPHost: getEnv("SMTP_HOST", "smtp.163.com"),
		SMTPPort: getEnv("SMTP_PORT", "465"),
		SMTPUser: getEnv("SMTP_USER", ""),
		SMTPPass: getEnv("SMTP_PASS", ""),
		SMTPFrom: getEnv("SMTP_FROM", getEnv("SMTP_USER", "")),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" { //get value from loaded env
		return value
	}
	return defaultValue
}
