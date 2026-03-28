package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"qrcode-gen/model"

	"github.com/redis/go-redis/v9"
)

type redisRepo struct {
	client *redis.Client
	ttl    time.Duration // 快取存活時間
	ctx    context.Context
}

func NewRedisRepository(addr string, ttl time.Duration) (Repository, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr, // 例如 "localhost:6379"
	})

	// 測試連線
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &redisRepo{
		client: client,
		ttl:    ttl,
		ctx:    ctx,
	}, nil
}

func (r *redisRepo) Create(qr *model.QRCode) error {
	data, err := json.Marshal(qr)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	// SET key value EX ttl
	// key 用 "qr:" 前綴，避免跟其他資料衝突
	key := "qr:" + qr.QRToken
	return r.client.Set(r.ctx, key, data, r.ttl).Err()
}

func (r *redisRepo) GetByToken(qrToken string) (*model.QRCode, error) {
	key := "qr:" + qrToken
	data, err := r.client.Get(r.ctx, key).Bytes()
	if err == redis.Nil {
		// key 不存在 = cache miss
		return nil, fmt.Errorf("qr code not found: %s", qrToken)
	}
	if err != nil {
		return nil, fmt.Errorf("redis get failed: %w", err)
	}

	var qr model.QRCode
	if err := json.Unmarshal(data, &qr); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}
	return &qr, nil
}

func (r *redisRepo) Delete(qrToken string) error {
	key := "qr:" + qrToken
	return r.client.Del(r.ctx, key).Err()
}

func (r *redisRepo) TokenExists(qrToken string) (bool, error) {
	key := "qr:" + qrToken
	count, err := r.client.Exists(r.ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists failed: %w", err)
	}
	return count > 0, nil
}

func (r *redisRepo) Update(qrToken string, url string) error {
	// cache 不做 update，交給 cached_repo 做 invalidation（刪除）
	return r.Delete(qrToken)
}

// 不適合用 cache 做，直接回傳空
func (r *redisRepo) GetByUserID(userID string) ([]*model.QRCode, error) {
	return nil, nil
}
