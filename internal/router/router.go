package router

import (
	"elk-stack-user/internal/handler"
	"elk-stack-user/internal/service"
	"elk-stack-user/internal/repository"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	// Repository ve service'leri oluştur
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	// Gin router'ı oluştur
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"message": "User service is running",
		})
	})

	// Login endpoint
	router.POST("/login", userHandler.Login)

	// User routes
	router.POST("/users", userHandler.CreateUser)
	router.GET("/users", userHandler.GetAllUsers)
	router.GET("/users/email", userHandler.GetUserByEmail)
	router.GET("/users/:id", userHandler.GetUserByID)
	router.PUT("/users/:id", userHandler.UpdateUser)
	router.DELETE("/users/:id", userHandler.DeleteUser)

	return router
}
