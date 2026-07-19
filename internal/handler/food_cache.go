package handler

import (
	"ai-calories/internal/model"
	"ai-calories/internal/service"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type FoodCacheHandler struct {
	cacheService *service.FoodCacheService
}

func NewFoodCacheHandler(cacheService *service.FoodCacheService) *FoodCacheHandler {
	return &FoodCacheHandler{cacheService: cacheService}
}

func (h *FoodCacheHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "query parameter 'q' is required"})
		return
	}

	results, err := h.cacheService.Search(q)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}
	if results == nil {
		results = []model.FoodCacheSearchResult{}
	}
	writeJSON(w, http.StatusOK, results)
}

func (h *FoodCacheHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "invalid id"})
		return
	}

	item, err := h.cacheService.GetByID(uint(id))
	if err != nil {
		writeJSON(w, http.StatusNotFound, model.ErrorResponse{Error: "food not found in cache"})
		return
	}
	writeJSON(w, http.StatusOK, item)
}
