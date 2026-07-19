package handler

import (
	"ai-calories/internal/model"
	"ai-calories/internal/service"
	"net/http"
)

type PaymentHandler struct {
	paymentService *service.PaymentService
}

func NewPaymentHandler(paymentService *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

func (h *PaymentHandler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	current, err := h.paymentService.GetCurrent(userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}
	if current == nil {
		writeJSON(w, http.StatusOK, nil)
		return
	}
	writeJSON(w, http.StatusOK, current)
}

func (h *PaymentHandler) Record(w http.ResponseWriter, r *http.Request) {
	var req model.RecordPaymentRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "invalid request body"})
		return
	}

	userID := getUserID(r)
	if err := h.paymentService.Record(userID, req); err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *PaymentHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	history, err := h.paymentService.GetHistory(userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}
	if history == nil {
		history = []model.PaymentHistoryItem{}
	}
	writeJSON(w, http.StatusOK, history)
}
