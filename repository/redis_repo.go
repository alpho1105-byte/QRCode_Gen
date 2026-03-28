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
	ttl    time.Duration
}

func NewRedisRepository(addr string, ttl time.Duration, poolSize int, minIdleConns int, readTimeout time.Duration, writeTimeout time.Duration, dialTimeout time.Duration) (Repository, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		PoolSize:     poolSize,
		MinIdleConns: minIdleConns,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		DialTimeout:  dialTimeout,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &redisRepo{
		client: client,
		ttl:    ttl,
	}, nil
}

func (r *redisRepo) Create(ctx context.Context, qr *model.QRCode) error {
	data, err := json.Marshal(qr)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	key := "qr:" + qr.QRToken
	return r.client.Set(ctx, key, data, r.ttl).Err()
}

func (r *redisRepo) GetByToken(ctx context.Context, qrToken string) (*model.QRCode, error) {
	key := "qr:" + qrToken
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
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

func (r *redisRepo) Delete(ctx context.Context, qrToken string) error {
	key := "qr:" + qrToken
	return r.client.Del(ctx, key).Err()
}

func (r *redisRepo) TokenExists(ctx context.Context, qrToken string) (bool, error) {
	key := "qr:" + qrToken
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists failed: %w", err)
	}
	return count > 0, nil
}

func (r *redisRepo) Update(ctx context.Context, qrToken string, url string) error {
	return r.Delete(ctx, qrToken)
}

func (r *redisRepo) GetByUserID(ctx context.Context, userID string) ([]*model.QRCode, error) {
	return nil, nil
}
