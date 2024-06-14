package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Pyegorchik/bdd/backend/internal/domain"
	"github.com/Pyegorchik/bdd/backend/models"
)

func (h *handler) Logout(w http.ResponseWriter, user *domain.UserWithTokenNumber, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.cfg.RequestTimeout)
	defer cancel()

	if err := h.service.Logout(ctx, user.ID, int64(user.Number), user.Role); err != nil {
		h.makeErrorResponse(w, r, err, code500)
		return
	}
	cookie := &http.Cookie{
		Name:     NameCookie,
		Value:    "",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "",
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
	refreshCookie := &http.Cookie{
		Name:     NameRefreshCookie,
		Value:    "",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/auth/refresh",
		MaxAge:   -1,
	}
	http.SetCookie(w, refreshCookie)
	result := true
	if err := writeResponse(w, r, http.StatusOK, &models.SuccessResponse{Success: &result}); err != nil {
		h.makeErrorResponse(w, r, err, code500)
		return
	}
}

func (h *handler) FullLogout(w http.ResponseWriter, user *domain.UserWithTokenNumber, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.cfg.RequestTimeout)
	defer cancel()

	if err := h.service.FullLogout(ctx, user.ID, user.Role); err != nil {
		h.makeErrorResponse(w, r, err, code500)
		return
	}
	cookie := &http.Cookie{
		Name:     NameCookie,
		Value:    "",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "",
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
	refreshCookie := &http.Cookie{
		Name:     NameRefreshCookie,
		Value:    "",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "",
		MaxAge:   -1,
	}
	http.SetCookie(w, refreshCookie)
	result := true
	if err := writeResponse(w, r, http.StatusOK, &models.SuccessResponse{Success: &result}); err != nil {
		h.makeErrorResponse(w, r, err, code500)
		return
	}
}

func (h *handler) AuthMessage(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req models.AuthMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.makeErrorResponse(w, r, err, code500)
		return
	}
	if err := req.Validate(h.validationFormats); err != nil {
		h.makeErrorResponse(w, r, makeValidationError("handleAuthMessage", err), code400)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), h.cfg.RequestTimeout)
	defer cancel()

	res, err := h.service.GetAuthMessage(ctx, &req)
	if err != nil {
		h.makeErrorResponse(w, r, err, code500)
		return
	}
	if err := writeResponse(w, r, http.StatusOK, res); err != nil {
		h.makeErrorResponse(w, r, err, code500)
		return
	}
}

func (h *handler) AuthByMessage(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req models.AuthBySignatureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.makeErrorResponse(w, r, err, code500)
		return
	}
	if err := req.Validate(h.validationFormats); err != nil {
		h.makeErrorResponse(w, r, makeValidationError("handleAuthByMessage", err), code400)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.cfg.RequestTimeout)
	defer cancel()

	res, accessToken, refreshToken, err := h.service.AuthByMessage(ctx, &req)
	if err != nil {
		h.makeErrorResponse(w, r, err, code500)
		return
	}

	cookie := &http.Cookie{
		Name:     NameCookie,
		Value:    accessToken.Token,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/api",
		MaxAge:   -int(time.Since(accessToken.ExpiresAt).Seconds()),
	}
	http.SetCookie(w, cookie)
	refreshCookie := &http.Cookie{
		Name:     NameRefreshCookie,
		Value:    refreshToken.Token,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/api/auth/refresh",
		MaxAge:   -int(time.Since(refreshToken.ExpiresAt).Seconds()),
	}
	http.SetCookie(w, refreshCookie)
	if err := writeResponse(w, r, http.StatusOK, res); err != nil {
		h.makeErrorResponse(w, r, err, code500)
		return
	}
}

func (h *handler) RefreshAuth(w http.ResponseWriter, user *domain.UserWithTokenNumber, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), h.cfg.RequestTimeout)
	defer cancel()

	res, accessToken, refreshToken, err := h.service.RefreshJWTokens(ctx, user.ID, int64(user.Number), user.Role)
	if err != nil {
		h.makeErrorResponse(w, r, err, code500)
		return
	}
	cookie := &http.Cookie{
		Name:     NameCookie,
		Value:    accessToken.Token,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/api",
		MaxAge:   -int(time.Since(accessToken.ExpiresAt).Seconds()),
	}
	http.SetCookie(w, cookie)
	refreshCookie := &http.Cookie{
		Name:     NameRefreshCookie,
		Value:    refreshToken.Token,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/api/auth/refresh",
		MaxAge:   -int(time.Since(refreshToken.ExpiresAt).Seconds()),
	}
	http.SetCookie(w, refreshCookie)
	if err := writeResponse(w, r, http.StatusOK, res); err != nil {
		h.makeErrorResponse(w, r, err, code500)
		return
	}
}
