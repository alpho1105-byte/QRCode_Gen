package config

import (
	"os"
	"time"
)

type Config struct {
	Port      string
	BaseURL   string
	RedisAddr string
	RedisTTL  time.Duration
}

func Load() *Config {
	return &Config{
		Port:      getEnv("PORT", "8080"),
		BaseURL:   getEnv("BASE_URL", "http://localhost:8080"),
		RedisAddr: getEnv("REDIS_ADDR", "localhost:6379"),
		RedisTTL:  parseDuration(getEnv("REDIS_TTL", "24h")),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 24 * time.Hour // 解析失敗用預設值
	}
	return d
}
