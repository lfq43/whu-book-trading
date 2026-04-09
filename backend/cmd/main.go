package main

import (
	"log"

	"book-trading/backend/internal/config"
	"book-trading/backend/internal/database"
	"book-trading/backend/internal/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	config.LoadConfig()

	// 初始化数据库
	database.InitMySQL()
	database.InitRedis()

	// 创建 Gin 引擎
	r := gin.Default()

	// 设置路由
	routes.SetupRoutes(r)

	// 健康检查（公开接口）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
	// 启动服务器
	log.Printf("Server starting on port %s", config.AppConfig.ServerPort)
	r.Run(":" + config.AppConfig.ServerPort)
}
