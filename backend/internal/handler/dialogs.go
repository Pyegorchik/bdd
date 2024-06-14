package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Pyegorchik/bdd/backend/internal/domain"
	"github.com/Pyegorchik/bdd/backend/models"
	"github.com/gorilla/mux"
)

func (h *handler) SendMessage(w http.ResponseWriter, user *domain.UserWithTokenNumber, r *http.Request) {
	defer r.Body.Close()
	var req models.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.makeErrorResponse(w, r, err, code500)
		return
	}
	if err := req.Validate(h.validationFormats); err != nil {
		h.makeErrorResponse(w, r, makeValidationError("handleSendMessage", err), code400)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), h.cfg.RequestTimeout)
	defer cancel()

	err := h.service.SendMessage(ctx, &req, user.ID)
	if err != nil {
		h.makeErrorResponse(w, r, err, code500)
		return
	}

	result := true
	if err := writeResponse(w, r, http.StatusOK, &models.SuccessResponse{Success: &result}); err != nil {
		h.makeErrorResponse(w, r, err, code500)
		return
	}
}

func (h *handler) GetDialogs(w http.ResponseWriter, user *domain.UserWithTokenNumber, r *http.Request) {
	defer r.Body.Close()

	ctx, cancel := context.WithTimeout(r.Context(), h.cfg.RequestTimeout)
	defer cancel()

	res, err := h.service.GetDialogs(ctx, user.ID)
	if err != nil {
		h.makeErrorResponse(w, r, err, code500)
		return
	}
	if err := writeResponse(w, r, http.StatusOK, res); err != nil {
		h.makeErrorResponse(w, r, err, code500)
		return
	}
}

func (h *handler) GetMessages(w http.ResponseWriter, user *domain.UserWithTokenNumber, r *http.Request) {
	defer r.Body.Close()
	dialogID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		h.makeErrorResponse(w, r, errors.New("invalid parameter value"), code400)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), h.cfg.RequestTimeout)
	defer cancel()

	res, err := h.service.GetMessages(ctx, int64(dialogID))
	if err != nil {
		h.makeErrorResponse(w, r, err, code500)
		return
	}
	if err := writeResponse(w, r, http.StatusOK, res); err != nil {
		h.makeErrorResponse(w, r, err, code500)
		return
	}
}
