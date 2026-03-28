package model

import (
	"time"
)

// QRCode 對應設計文件中的 QrCodes table
// 欄位：id, user_id, qr_token, url, created_at
type QRCode struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	QRToken   string    `json:"qr_token"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateRequest struct {
	URL string `json:"url"`
}

type CreateResponse struct {
	QRToken string `json:"qr_token"`
}

type URLResponse struct {
	URL string `json:"url"`
}

type ImageResponse struct {
	ImageLocation string `json:"image_location"`
}
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
