package logger

import (
	"time"
	"github.com/gin-gonic/gin"
)

// RequestIDMiddleware request ID ekler
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// LoggingMiddleware HTTP request'leri loglar
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Request başlangıcı
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		requestID := c.GetString("request_id")

		// Request log'u
		Logger.Info("HTTP Request Started",
			RequestID(requestID),
			Method(method),
			Path(path),
			ClientIP(clientIP),
			UserAgent(userAgent),
			Query(raw),
		)

		// Request'i işle
		c.Next()

		// Response bilgileri
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		// Response log'u
		Logger.Info("HTTP Request Completed",
			RequestID(requestID),
			Method(method),
			Path(path),
			StatusCode(statusCode),
			ResponseTime(latency),
			Int("body_size", bodySize),
			ClientIP(clientIP),
		)

		// Error log'u (4xx, 5xx status kodları için)
		if statusCode >= 400 {
			Logger.Error("HTTP Request Error",
				RequestID(requestID),
				Method(method),
				Path(path),
				StatusCode(statusCode),
				ResponseTime(latency),
				ClientIP(clientIP),
				String("error_message", c.Errors.String()),
			)
		}
	}
}

// generateRequestID basit bir request ID oluşturur
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString rastgele string oluşturur
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// ErrorLogger error'ları loglar
func ErrorLogger() gin.HandlerFunc {
	return gin.ErrorLogger()
}

// RecoveryLogger panic'leri yakalar ve loglar
func RecoveryLogger() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		requestID := c.GetString("request_id")
		clientIP := c.ClientIP()
		path := c.Request.URL.Path
		method := c.Request.Method

		Logger.Error("Panic Recovered",
			RequestID(requestID),
			Method(method),
			Path(path),
			ClientIP(clientIP),
			Any("panic", recovered),
			String("stack", getStackTrace()),
		)

		c.AbortWithStatus(500)
	})
}

// getStackTrace stack trace alır (basit implementasyon)
func getStackTrace() string {
	// Gerçek uygulamada runtime/debug.Stack() kullanılabilir
	return "Stack trace not implemented"
}
