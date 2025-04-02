package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/yourusername/gin-vue-chat/config"
	"github.com/yourusername/gin-vue-chat/controllers"
	"github.com/yourusername/gin-vue-chat/middlewares"
	"github.com/yourusername/gin-vue-chat/models"
	"github.com/yourusername/gin-vue-chat/websocket"
)

func main() {
	// 初始化配置
	config.InitConfig()

	// 初始化数据库
	models.InitDB()

	// 初始化MongoDB
	models.InitMongoDB()

	// 创建Gin实例
	r := gin.Default()

	// 配置CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 初始化WebSocket管理器
	hub := websocket.NewHub()
	go hub.Run()

	// 将WebSocket Hub添加到Gin上下文中
	r.Use(func(c *gin.Context) {
		c.Set("wsHub", hub)
		c.Next()
	})

	// 公共路由组
	public := r.Group("/api")
	{
		// 认证相关路由
		auth := public.Group("/auth")
		{
			auth.POST("/register", controllers.Register)
			auth.POST("/login", controllers.Login)
		}
	}

	// 需要认证的路由组
	protected := r.Group("/api")
	protected.Use(middlewares.JWTAuth())
	{
		// 用户相关路由
		user := protected.Group("/user")
		{
			user.GET("/profile", controllers.GetUserProfile)
			user.PUT("/profile", controllers.UpdateUserProfile)
			user.PUT("/password", controllers.ChangePassword)
		}

		// 好友相关路由
		friends := protected.Group("/friends")
		{
			friends.GET("", controllers.GetFriends)
			friends.POST("/add", controllers.AddFriend)
			friends.DELETE("/:id", controllers.RemoveFriend)
		}

		// 群组相关路由
		groups := protected.Group("/groups")
		{
			groups.GET("", controllers.GetGroups)
			groups.POST("/create", controllers.CreateGroup)
			groups.GET("/:id", controllers.GetGroupDetail)
			groups.PUT("/:id", controllers.UpdateGroup)
			groups.DELETE("/:id", controllers.DeleteGroup)
			groups.GET("/:id/members", controllers.GetGroupMembers)
			groups.POST("/:id/members", controllers.AddGroupMember)
			groups.DELETE("/:id/members/:userId", controllers.RemoveGroupMember)
		}

		// 消息相关路由
		messages := protected.Group("/messages")
		{
			messages.GET("/private/:userId", controllers.GetPrivateMessages)
			messages.POST("/private", controllers.SendPrivateMessage)
			messages.GET("/group/:groupId", controllers.GetGroupMessages)
			messages.POST("/group", controllers.SendGroupMessage)
		}
	}

	// WebSocket路由
	r.GET("/ws", middlewares.JWTAuth(), func(c *gin.Context) {
		websocket.ServeWs(hub, c)
	})

	// 启动服务器
	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	log.Println("Server is running on http://localhost:8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}
