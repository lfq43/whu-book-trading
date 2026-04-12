package routes

import (
	"book-trading/backend/internal/controllers"
	"book-trading/backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		// 公开路由
		auth := api.Group("/auth")
		{
			auth.POST("/register", controllers.Register)
			auth.POST("/login", controllers.Login)
		}

		// 用户公开路由（查看个人空间）
		api.GET("/users/:id", controllers.GetUserProfile)

		// WebSocket 聊天
		api.GET("/ws", controllers.Websocket)

		// 批次公开路由
		api.GET("/batches", controllers.GetBatchList)
		api.GET("/batches/:id", controllers.GetBatchDetail)

		// 需要认证的路由
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// 用户相关
			user := protected.Group("/user")
			{
				user.GET("/profile", controllers.GetProfile)
				user.PUT("/profile", controllers.UpdateProfile)
				user.GET("/batches", controllers.GetMyBatches)
			}

			// 批次管理
			batch := protected.Group("/batches")
			{
				batch.POST("", controllers.CreateBatch)
				batch.PUT("/:id/image", controllers.UpdateBatchImage)
				batch.PUT("/:id/book-status", controllers.UpdateBookSoldStatus)
				batch.DELETE("/:id", controllers.DeleteBatch)
			}

			// 消息相关
			message := protected.Group("/messages")
			{
				message.POST("", controllers.SendMessage)                         // 发送消息
				message.GET("/unread", controllers.GetUnreadCount)                // 未读数量
				message.GET("/conversations", controllers.GetConversationList)    // 对话列表
				message.GET("/conversation/:userId", controllers.GetConversation) // 与某人的聊天记录
			}

			// 管理员相关
			admin := protected.Group("/admin")
			admin.Use(middleware.AdminMiddleware())
			{
				admin.GET("/users", controllers.GetAllUsers)
				admin.PUT("/users/:id/ban", controllers.BanUser)
			}

			// 文件相关
			upload := protected.Group("/upload")
			{
				upload.POST("/avatar", controllers.UploadAvatar)
				upload.POST("/image", controllers.UploadImage)
			}
		}
	}
}
