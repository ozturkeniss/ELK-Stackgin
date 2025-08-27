package logger

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// ELKSender Logstash'e TCP ile log gönderir
type ELKSender struct {
	conn net.Conn
	addr string
}

// NewELKSender yeni ELK sender oluşturur
func NewELKSender(addr string) (*ELKSender, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to logstash: %w", err)
	}

	return &ELKSender{
		conn: conn,
		addr: addr,
	}, nil
}

// SendLog log'u Logstash'e gönderir
func (s *ELKSender) SendLog(level, message string, fields map[string]interface{}) error {
	logEntry := map[string]interface{}{
		"timestamp":   time.Now().Format(time.RFC3339),
		"level":       level,
		"message":     message,
		"service":     "user-service",
		"environment": "development",
	}

	// Fields'ları ekle
	for k, v := range fields {
		logEntry[k] = v
	}

	// JSON'a çevir
	data, err := json.Marshal(logEntry)
	if err != nil {
		return fmt.Errorf("failed to marshal log: %w", err)
	}

	// Newline ekle (Logstash için gerekli)
	data = append(data, '\n')

	// Gönder
	_, err = s.conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to send log: %w", err)
	}

	return nil
}

// Close bağlantıyı kapatır
func (s *ELKSender) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}

// Reconnect bağlantıyı yeniden kurar
func (s *ELKSender) Reconnect() error {
	if s.conn != nil {
		s.conn.Close()
	}

	conn, err := net.Dial("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to reconnect: %w", err)
	}

	s.conn = conn
	return nil
}
