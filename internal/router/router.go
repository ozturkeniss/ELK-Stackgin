package router

import (
	"elk-stack-user/internal/handler"
	"elk-stack-user/internal/service"
	"elk-stack-user/internal/repository"
	"elk-stack-user/internal/logger"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	// Repository ve service'leri oluştur
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	// Gin router'ı oluştur
	router := gin.New()

	// Middleware'ler - Zap logger kullan
	router.Use(logger.RequestIDMiddleware())
	router.Use(logger.LoggingMiddleware())
	router.Use(logger.RecoveryLogger())
	router.Use(CORSMiddleware())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"message": "User service is running",
		})
	})

	// API routes
	api := router.Group("")
	{
		// User routes
		users := api.Group("/users")
		{
			users.POST("", userHandler.CreateUser)           // POST /users
			users.GET("", userHandler.GetAllUsers)          // GET /users
			users.GET("/email", userHandler.GetUserByEmail) // GET /users/email?email=...
			users.GET("/:id", userHandler.GetUserByID)      // GET /users/:id
			users.PUT("/:id", userHandler.UpdateUser)       // PUT /users/:id
			users.DELETE("/:id", userHandler.DeleteUser)    // DELETE /users/:id
		}
	}

	return router
}

// CORSMiddleware CORS ayarları
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
