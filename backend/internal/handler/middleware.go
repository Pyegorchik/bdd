package handler

import (
	"context"
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

func (h *handler) CookieAuthMiddleware(next http.Handler) http.Handler {
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
		ctx := context.WithValue(r.Context(), CtxKeyUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *handler) CookieRefreshAuthMiddleware(next http.Handler) http.Handler {
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
		ctx := context.WithValue(r.Context(), CtxKeyUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *handler) rolePermissionMiddleware(role domain.Role, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userContext, ok := r.Context().Value(CtxKeyUser).(domain.UserWithTokenNumber)
		if !ok {
			if err := writeResponse(w, r, http.StatusBadRequest, models.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("Invalid UserChain with role %w", role),
			}); err != nil {
				h.logging.Error(fmt.Errorf("write response: %w", err))
			}

			return
		}

		if role == userContext.Role {
			ctx := context.WithValue(r.Context(), CtxKeyUser, userContext)
			next.ServeHTTP(w, r.WithContext(ctx))

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
