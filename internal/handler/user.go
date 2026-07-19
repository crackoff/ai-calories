package handler

import (
	"ai-calories/internal/model"
	"ai-calories/internal/service"
	"net/http"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	profile, err := h.userService.GetProfile(userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

func (h *UserHandler) UpdateTimezone(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateTimezoneRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "invalid request body"})
		return
	}

	userID := int64(getUserID(r))
	if err := h.userService.UpdateTimezone(userID, req.Timezone); err != nil {
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) UpdateLanguage(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateLanguageRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "invalid request body"})
		return
	}

	userID := getUserID(r)
	if err := h.userService.UpdateLanguage(userID, req.Language); err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
