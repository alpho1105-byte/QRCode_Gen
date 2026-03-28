package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"qrcode-gen/model"
	"qrcode-gen/service"

	"github.com/go-chi/chi/v5"
	qrcode "github.com/skip2/go-qrcode"
)

type Handler struct {
	svc                service.Service
	baseURL            string
	defaultQRDimension int
	maxQRDimension     int
	cacheMaxAge        int
}

func NewHandler(svc service.Service, baseURL string, defaultQRDimension int, maxQRDimension int, cacheMaxAge int) *Handler {
	return &Handler{
		svc:                svc,
		baseURL:            baseURL,
		defaultQRDimension: defaultQRDimension,
		maxQRDimension:     maxQRDimension,
		cacheMaxAge:        cacheMaxAge,
	}
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
	ctx := r.Context()
	var req model.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Request body must be valid JSON with a 'url' field")
		return
	}

	userID := "demo-user"
	resp, err := h.svc.CreateQRCode(ctx, userID, req.URL)
	if err != nil {
		writeError(w, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) GetQRCodeImage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	qrToken := chi.URLParam(r, "qr_token")

	_, err := h.svc.GetOriginalURL(ctx, qrToken)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "QR code not found")
		return
	}

	dimension := h.defaultQRDimension
	if d := r.URL.Query().Get("dimension"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= h.maxQRDimension {
			dimension = parsed
		}
	}

	redirectURL := h.baseURL + "/r/" + qrToken
	png, err := qrcode.Encode(redirectURL, qrcode.Medium, dimension)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "QR_GEN_FAILED", "Failed to generate QR code image")
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", h.cacheMaxAge))
	w.WriteHeader(http.StatusOK)
	w.Write(png)
}

func (h *Handler) GetOriginalURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	qrToken := chi.URLParam(r, "qr_token")

	originalURL, err := h.svc.GetOriginalURL(ctx, qrToken)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "QR code not found")
		return
	}

	writeJSON(w, http.StatusOK, model.URLResponse{URL: originalURL})
}

func (h *Handler) UpdateQRCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	qrToken := chi.URLParam(r, "qr_token")

	var req model.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Request body must contain a 'url' field")
		return
	}

	if err := h.svc.UpdateQRCode(ctx, qrToken, req.URL); err != nil {
		writeError(w, http.StatusNotFound, "UPDATE_FAILED", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteQRCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	qrToken := chi.URLParam(r, "qr_token")

	if err := h.svc.DeleteQRCode(ctx, qrToken); err != nil {
		writeError(w, http.StatusNotFound, "DELETE_FAILED", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	qrToken := chi.URLParam(r, "qr_token")

 	originalURL, err := h.svc.GetOriginalURL(ctx, qrToken)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "QR code not found")
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, errCode string, message string) {
	writeJSON(w, status, model.ErrorResponse{
		Error:   errCode,
		Message: message,
	})
}
