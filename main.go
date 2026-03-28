package main

/*
	 HTTP Request
     	 ↓
┌─────────────────────┐
│   main.go           │  組裝所有元件，啟動 server
│   (組裝 & 啟動)      │
└────────┬────────────┘
         ↓
┌─────────────────────┐
│   middleware/        │  每個 request 先經過這裡（logging, auth...）
│   (攔截 & 紀錄)      │
└────────┬────────────┘
         ↓
┌─────────────────────┐
│   handler/           │  解析 HTTP → 呼叫 Service → 回傳 HTTP
│   (HTTP 翻譯層)      │  只管「怎麼接、怎麼回」
└────────┬────────────┘
         ↓
┌─────────────────────┐
│   service/           │  商業邏輯：驗證 URL、產生 token、重試 collision
│   (核心業務)         │  不知道也不在乎 HTTP 的存在
└────────┬────────────┘
         ↓
┌──────────┴──────────┐
│ repository/    token/ │
│ (存取資料)   (產生token)│
└─────────────────────┘
*/

import (
	"fmt"
	"log"
	"net/http"

	"qrcode-gen/config"
	"qrcode-gen/handler"
	"qrcode-gen/middleware"
	"qrcode-gen/repository"
	"qrcode-gen/service"
	"qrcode-gen/token"

	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.Load()

	sqliteDB, err := repository.NewSQLiteRepository(
		cfg.DBPath,
		cfg.DBMaxOpenConns,
		cfg.DBMaxIdleConns,
		cfg.DBConnMaxLifetime,
	)
	if err != nil {
		log.Fatalf("failed to init database: %v", err)
	}

	db := repository.NewBreakerRepository(
		sqliteDB,
		cfg.BreakerMaxRequests,
		cfg.BreakerInterval,
		cfg.BreakerTimeout,
		cfg.BreakerFailThreshold,
	)

	cache, err := repository.NewRedisRepository(
		cfg.RedisAddr,
		cfg.RedisTTL,
		cfg.RedisPoolSize,
		cfg.RedisMinIdleConns,
		cfg.RedisReadTimeout,
		cfg.RedisWriteTimeout,
		cfg.RedisDialTimeout,
	)
	if err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}

	repo := repository.NewCompositeRepository(cache, db)

	tokenGen := token.NewGenerator(cfg.TokenLength)

	svc := service.NewService(repo, tokenGen, cfg.BaseURL, cfg.MaxRetries, cfg.MaxURLLength)

	h := handler.NewHandler(svc, cfg.BaseURL, cfg.DefaultQRDimension, cfg.MaxQRDimension, cfg.CacheMaxAge)

	r := chi.NewRouter()
	r.Use(middleware.Logging)
	h.RegisterRoutes(r)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "OK")
	})

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("QR Code Generator starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
