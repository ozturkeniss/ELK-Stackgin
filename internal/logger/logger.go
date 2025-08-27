package logger

import (
	"os"
	"time"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger     *zap.Logger
	elkSender  *ELKSender
)

// InitLogger logger'ı başlatır
func InitLogger(environment string) error {
	var config zap.Config

	if environment == "production" {
		// Production: JSON format, INFO ve üzeri seviyeler
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		
		// Dosyaya yazma
		config.OutputPaths = []string{"logs/app.log", "stdout"}
		config.ErrorOutputPaths = []string{"logs/error.log", "stderr"}
	} else {
		// Development: Console format, DEBUG ve üzeri seviyeler
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
		config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}

	// Logger'ı oluştur
	var err error
	Logger, err = config.Build()
	if err != nil {
		return err
	}

	// Global logger'ı ayarla
	zap.ReplaceGlobals(Logger)

	// ELK Sender'ı başlat (opsiyonel)
	if elkAddr := os.Getenv("ELK_LOGSTASH_ADDR"); elkAddr != "" {
		sender, err := NewELKSender(elkAddr)
		if err != nil {
			Logger.Warn("Failed to connect to ELK Stack", zap.Error(err))
		} else {
			elkSender = sender
			Logger.Info("Connected to ELK Stack", zap.String("address", elkAddr))
		}
	}

	return nil
}

// Sync logger'ı kapatır
func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}

// LogLevel string'den log level'a çevirir
func LogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// CreateLogsDirectory logs dizinini oluşturur
func CreateLogsDirectory() error {
	return os.MkdirAll("logs", 0755)
}

// LogFields için yardımcı fonksiyonlar
func String(key, val string) zap.Field {
	return zap.String(key, val)
}

func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

func Uint(key string, val uint) zap.Field {
	return zap.Uint(key, val)
}

func Float64(key string, val float64) zap.Field {
	return zap.Float64(key, val)
}

func Bool(key string, val bool) zap.Field {
	return zap.Bool(key, val)
}

func Duration(key string, val time.Duration) zap.Field {
	return zap.Duration(key, val)
}

func Time(key string, val time.Time) zap.Field {
	return zap.Time(key, val)
}

func Any(key string, val interface{}) zap.Field {
	return zap.Any(key, val)
}

func Error(err error) zap.Field {
	return zap.Error(err)
}

// HTTP Request için özel field'lar
func RequestID(id string) zap.Field {
	return zap.String("request_id", id)
}

func ClientIP(ip string) zap.Field {
	return zap.String("client_ip", ip)
}

func UserAgent(ua string) zap.Field {
	return zap.String("user_agent", ua)
}

func Method(method string) zap.Field {
	return zap.String("method", method)
}

func Path(path string) zap.Field {
	return zap.String("path", path)
}

func StatusCode(code int) zap.Field {
	return zap.Int("status_code", code)
}

func ResponseTime(duration time.Duration) zap.Field {
	return zap.Duration("response_time", duration)
}

// Database için özel field'lar
func Query(query string) zap.Field {
	return zap.String("query", query)
}

func Table(table string) zap.Field {
	return zap.String("table", table)
}

func RowsAffected(count int64) zap.Field {
	return zap.Int64("rows_affected", count)
}

// Business logic için özel field'lar
func UserID(id uint) zap.Field {
	return zap.Uint("user_id", id)
}

func Username(username string) zap.Field {
	return zap.String("username", username)
}

func Email(email string) zap.Field {
	return zap.String("email", email)
}

func ServiceName(name string) zap.Field {
	return zap.String("service", name)
}
