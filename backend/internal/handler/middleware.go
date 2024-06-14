package handler

import (
	"errors"
	"fmt"

	"net/http"

	"github.com/Pyegorchik/bdd/backend/internal/domain"
	"github.com/Pyegorchik/bdd/backend/models"
	"github.com/Pyegorchik/bdd/backend/pkg/jwtoken"
)

type CtxKey string

const (
	CtxKeyUser        = CtxKey("user")
	TokenStart        = "Bearer "
	NameCookie        = "access-token"
	NameRefreshCookie = "refresh-token"
)

func (h *handler) CookieAuthMiddleware(next HandlerFuncWithUser) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			cookie *http.Cookie
			err    error
		)
		cookie, err = r.Cookie(NameCookie)
		if err != nil || cookie == nil {
			h.logging.Info("cookie not found in ws connect")
			r.Body.Close()
			h.makeErrorResponse(w, r, errors.New("missing credentials"), code401)
			return
		}
		token := cookie.Value
		user, err := h.service.GetUserByJWToken(r.Context(), jwtoken.PurposeAccess, token)
		if err != nil {
			r.Body.Close()
			h.makeErrorResponse(w, r, err, code500)
			return
		}

		next(w, user, r)
	})
}

func (h *handler) UnnecessaryCookieAuthMiddleware(next HandlerFuncWithUser) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			cookie *http.Cookie
			err    error
		)
		cookie, err = r.Cookie(NameCookie)
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				next(w, nil, r)
				return
			}
			h.logging.Info("cookie not found in ws connect")
			r.Body.Close()
			h.makeErrorResponse(w, r, errors.New("missing credentials"), code500)
			return
		}

		token := cookie.Value
		user, err := h.service.GetUserByJWToken(r.Context(), jwtoken.PurposeAccess, token)
		if err != nil {
			r.Body.Close()
			h.makeErrorResponse(w, r, err, code500)
			return
		}

		next(w, user, r)
	})
}

func (h *handler) CookieRefreshAuthMiddleware(next HandlerFuncWithUser) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			cookie *http.Cookie
			err    error
		)

		cookie, err = r.Cookie(NameRefreshCookie)
		if err != nil || cookie == nil {
			h.logging.Info("cookie not found in ws connect")
			r.Body.Close()
			h.makeErrorResponse(w, r, errors.New("missing credentials"), code401)
			return
		}
		token := cookie.Value
		user, err := h.service.GetUserByJWToken(r.Context(), jwtoken.PurposeRefresh, token)
		if err != nil {
			r.Body.Close()
			h.makeErrorResponse(w, r, err, code500)
			return
		}

		next(w, user, r)
	})
}

func (h *handler) rolePermissionMiddleware(role domain.Role, next HandlerFuncWithUser) HandlerFuncWithUser {
	return HandlerFuncWithUser(func(w http.ResponseWriter, user *domain.UserWithTokenNumber, r *http.Request) {
		if role == user.Role {
			next(w, user, r)
			return
		}

		if err := writeResponse(w, r, http.StatusForbidden, models.ErrorResponse{
			Code:    http.StatusForbidden,
			Message: "permissionDenied",
		}); err != nil {
			h.logging.Error(fmt.Errorf("write response: %w", err))
		}
	})
}
