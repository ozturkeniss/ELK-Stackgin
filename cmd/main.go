package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"elk-stack-user/internal/database"
	"elk-stack-user/internal/router"
	"elk-stack-user/internal/logger"
)

func main() {
	// Environment
	env := getEnv("ENV", "development")
	
	// Logger'ı başlat
	if err := logger.InitLogger(env); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	// Logs dizinini oluştur
	if err := logger.CreateLogsDirectory(); err != nil {
		logger.Logger.Error("Failed to create logs directory", logger.Error(err))
	}

	logger.Logger.Info("Starting User Service", 
		logger.String("environment", env),
		logger.String("version", "1.0.0"),
	)

	// Database configuration
	dbConfig := database.NewConfig()
	
	// Database connection
	db, err := database.Connect(dbConfig)
	if err != nil {
		logger.Logger.Fatal("Failed to connect to database", logger.Error(err))
	}
	defer database.Close(db)

	// Auto migrate database
	if err := database.AutoMigrate(db); err != nil {
		logger.Logger.Fatal("Failed to migrate database", logger.Error(err))
	}

	// Setup router
	r := router.SetupRouter(db)

	// Server configuration
	port := getEnv("PORT", "8080")
	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Graceful shutdown için channel
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Server'ı background'da başlat
	go func() {
		logger.Logger.Info("Server starting", 
			logger.String("port", port),
			logger.String("address", ":"+port),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Fatal("Failed to start server", logger.Error(err))
		}
	}()

	// Graceful shutdown
	<-quit
	logger.Logger.Info("Server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Logger.Fatal("Server forced to shutdown", logger.Error(err))
	}

	logger.Logger.Info("Server exited")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
