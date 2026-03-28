package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Server
	Port    string
	BaseURL string

	// Redis
	RedisAddr         string
	RedisTTL          time.Duration
	RedisPoolSize     int
	RedisMinIdleConns int
	RedisReadTimeout  time.Duration
	RedisWriteTimeout time.Duration
	RedisDialTimeout  time.Duration

	// SQLite
	DBPath            string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration

	// Circuit Breaker
	BreakerMaxRequests   uint32        // 半開狀態最多放幾個 request 試探
	BreakerInterval      time.Duration // 錯誤計數重置間隔
	BreakerTimeout       time.Duration // 熔斷後等多久再試
	BreakerFailThreshold uint32        // 連續錯誤幾次觸發熔斷

	// Token
	TokenLength int // Base62 token 取幾個字元
	MaxRetries  int // 碰撞時最多重試幾次

	// QR Code
	MaxURLLength       int // URL 最大長度
	DefaultQRDimension int // QR code 預設尺寸
	MaxQRDimension     int // QR code 最大尺寸
	CacheMaxAge        int // Cache-Control max-age（秒）
}

func Load() *Config {
	return &Config{
		Port:    getEnv("PORT", "8080"),
		BaseURL: getEnv("BASE_URL", "http://localhost:8080"),

		RedisAddr:         getEnv("REDIS_ADDR", "localhost:6379"),
		RedisTTL:          parseDuration(getEnv("REDIS_TTL", "24h")),
		RedisPoolSize:     getEnvInt("REDIS_POOL_SIZE", 20),
		RedisMinIdleConns: getEnvInt("REDIS_MIN_IDLE_CONNS", 5),
		RedisReadTimeout:  parseDuration(getEnv("REDIS_READ_TIMEOUT", "2s")),
		RedisWriteTimeout: parseDuration(getEnv("REDIS_WRITE_TIMEOUT", "2s")),
		RedisDialTimeout:  parseDuration(getEnv("REDIS_DIAL_TIMEOUT", "3s")),

		DBPath:            getEnv("DB_PATH", "qrcode.db"),
		DBMaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
		DBConnMaxLifetime: parseDuration(getEnv("DB_CONN_MAX_LIFETIME", "5m")),

		BreakerMaxRequests:   uint32(getEnvInt("BREAKER_MAX_REQUESTS", 3)),
		BreakerInterval:      parseDuration(getEnv("BREAKER_INTERVAL", "10s")),
		BreakerTimeout:       parseDuration(getEnv("BREAKER_TIMEOUT", "5s")),
		BreakerFailThreshold: uint32(getEnvInt("BREAKER_FAIL_THRESHOLD", 5)),

		TokenLength: getEnvInt("TOKEN_LENGTH", 8),
		MaxRetries:  getEnvInt("TOKEN_MAX_RETRIES", 5),

		MaxURLLength:       getEnvInt("MAX_URL_LENGTH", 20),
		DefaultQRDimension: getEnvInt("DEFAULT_QR_DIMENSION", 256),
		MaxQRDimension:     getEnvInt("MAX_QR_DIMENSION", 1024),
		CacheMaxAge:        getEnvInt("CACHE_MAX_AGE", 86400),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 24 * time.Hour
	}
	return d
}
