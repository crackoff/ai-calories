package handler

import (
	"ai-calories/internal/model"
	"ai-calories/internal/service"
	"net/http"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "invalid request body"})
		return
	}
	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "email and password are required"})
		return
	}

	resp, err := h.authService.Register(req.Email, req.Password)
	if err != nil {
		writeJSON(w, http.StatusConflict, model.ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "invalid request body"})
		return
	}

	resp, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, model.ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	var req model.OAuthRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "invalid request body"})
		return
	}

	resp, err := h.authService.GoogleLogin(req.IDToken)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, model.ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) AppleLogin(w http.ResponseWriter, r *http.Request) {
	var req model.OAuthRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "invalid request body"})
		return
	}

	resp, err := h.authService.AppleLogin(req.IDToken)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, model.ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req model.RefreshRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "invalid request body"})
		return
	}

	resp, err := h.authService.Refresh(req.RefreshToken)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, model.ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}
