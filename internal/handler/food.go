package handler

import (
	"ai-calories/internal/model"
	"ai-calories/internal/service"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type FoodHandler struct {
	foodService *service.FoodService
}

func NewFoodHandler(foodService *service.FoodService) *FoodHandler {
	return &FoodHandler{foodService: foodService}
}

func (h *FoodHandler) LogFood(w http.ResponseWriter, r *http.Request) {
	var req model.LogFoodRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "invalid request body"})
		return
	}

	userID := int64(getUserID(r))
	resp, err := h.foodService.LogFood(userID, req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *FoodHandler) GetTodayFoods(w http.ResponseWriter, r *http.Request) {
	userID := int64(getUserID(r))
	foods, err := h.foodService.GetFoodsByDate(userID, time.Now())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, foods)
}

func (h *FoodHandler) GetFoodsByDate(w http.ResponseWriter, r *http.Request) {
	dateStr := chi.URLParam(r, "date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "invalid date format, use YYYY-MM-DD"})
		return
	}

	userID := int64(getUserID(r))
	foods, err := h.foodService.GetFoodsByDate(userID, date)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, foods)
}

func (h *FoodHandler) GetTodaySummary(w http.ResponseWriter, r *http.Request) {
	userID := int64(getUserID(r))
	summary, err := h.foodService.GetTodaySummary(userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (h *FoodHandler) GetDateSummary(w http.ResponseWriter, r *http.Request) {
	dateStr := chi.URLParam(r, "date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "invalid date format, use YYYY-MM-DD"})
		return
	}

	userID := int64(getUserID(r))
	summary, err := h.foodService.GetDateSummary(userID, date)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (h *FoodHandler) GetFoodHistory(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "week"
	}

	userID := int64(getUserID(r))
	history, err := h.foodService.GetFoodHistory(userID, period)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, history)
}

func (h *FoodHandler) DeleteLast(w http.ResponseWriter, r *http.Request) {
	userID := int64(getUserID(r))
	if err := h.foodService.DeleteLastFood(userID); err != nil {
		writeJSON(w, http.StatusNotFound, model.ErrorResponse{Error: "no food entry found"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *FoodHandler) DeleteByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "invalid id"})
		return
	}

	userID := int64(getUserID(r))
	if err := h.foodService.DeleteFood(userID, uint(id)); err != nil {
		writeJSON(w, http.StatusNotFound, model.ErrorResponse{Error: "food entry not found"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
