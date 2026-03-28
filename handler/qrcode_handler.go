package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"qrcode-gen/model"
	"qrcode-gen/service"
	"strconv"

	"github.com/go-chi/chi/v5"
	qrcode "github.com/skip2/go-qrcode"
)

type Handler struct {
	svc     service.Service
	baseURL string
}

func NewHandler(svc service.Service, baseURL string) *Handler {
	return &Handler{svc: svc, baseURL: baseURL}
}
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/v1", func(r chi.Router) {
		r.Post("/qr_code", h.CreateQRCode)
		r.Get("/qr_code/{qr_token}", h.GetOriginalURL)
		r.Get("/qr_code_image/{qr_token}", h.GetQRCodeImage)
		r.Put("/qr_code/{qr_token}", h.UpdateQRCode)
		r.Delete("/qr_code/{qr_token}", h.DeleteQRCode)
	})

	r.Get("/r/{qr_token}", h.Redirect)
}

func (h *Handler) CreateQRCode(w http.ResponseWriter, r *http.Request) {
	// 1. 從 request body 讀取 JSON
	var req model.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Request body must be valid JSON with a 'url' field")
		return
	}

	// 2. 呼叫 service（暫時用固定 userID，之後加認證再改）
	userID := "demo-user"
	resp, err := h.svc.CreateQRCode(userID, req.URL)
	if err != nil {
		writeError(w, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}

	// 3. 回傳 201 Created
	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) GetOriginalURL(w http.ResponseWriter, r *http.Request) {
	// 從路徑取出 qr_token（chi 的功能）
	qrToken := chi.URLParam(r, "qr_token")

	originalURL, err := h.svc.GetOriginalURL(qrToken)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "QR code not found")
		return
	}

	writeJSON(w, http.StatusOK, model.URLResponse{URL: originalURL})
}

func (h *Handler) UpdateQRCode(w http.ResponseWriter, r *http.Request) {
	qrToken := chi.URLParam(r, "qr_token")

	var req model.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Request body must contain a 'url' field")
		return
	}

	if err := h.svc.UpdateQRCode(qrToken, req.URL); err != nil {
		writeError(w, http.StatusNotFound, "UPDATE_FAILED", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204：成功但沒有內容回傳
}

func (h *Handler) DeleteQRCode(w http.ResponseWriter, r *http.Request) {
	qrToken := chi.URLParam(r, "qr_token")

	if err := h.svc.DeleteQRCode(qrToken); err != nil {
		writeError(w, http.StatusNotFound, "DELETE_FAILED", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- 工具函數：統一 JSON 回傳格式 ---

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	qrToken := chi.URLParam(r, "qr_token")

	originalURL, err := h.svc.GetOriginalURL(qrToken)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "QR code not found")
		return
	}

	// 設計文件：選 302 而非 301
	// 原因：確保每次都經過我們的 server，讓擁有者可以修改或刪除
	http.Redirect(w, r, originalURL, http.StatusFound)
}

func (h *Handler) GetQRCodeImage(w http.ResponseWriter, r *http.Request) {
	qrToken := chi.URLParam(r, "qr_token")

	// 1. 確認 token 存在
	_, err := h.svc.GetOriginalURL(qrToken)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "QR code not found")
		return
	}

	// 2. 解析 query parameter：dimension（圖片尺寸）
	dimension := 256 // 預設 256x256
	if d := r.URL.Query().Get("dimension"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= 1024 {
			dimension = parsed
		}
	}

	// 設計文件：「QR Code 內嵌的 URL 會是 https://myqrcode.com/qr_token」
	redirectURL := h.baseURL + "/r/" + qrToken
	png, err := qrcode.Encode(redirectURL, qrcode.Medium, dimension)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "QR_GEN_FAILED", "Failed to generate QR code image")
		return
	}

	// 4. 回傳 PNG 圖片
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(http.StatusOK)
	w.Write(png)
}

func writeError(w http.ResponseWriter, status int, errCode string, message string) {
	writeJSON(w, status, model.ErrorResponse{
		Error:   errCode,
		Message: message,
	})
}
