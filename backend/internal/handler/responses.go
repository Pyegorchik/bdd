package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Pyegorchik/bdd/backend/internal/service"
	"github.com/Pyegorchik/bdd/backend/models"
)

const (
	code400 = http.StatusBadRequest
	code401 = http.StatusUnauthorized
	code500 = http.StatusInternalServerError
	invalidUser          = "Invalid user"
)

var (
	UserMissingInCtxErr = errors.New("user is missing in context")
)

func (h *handler) makeErrorResponse(w http.ResponseWriter, r *http.Request, err error, code int) {
	var errService *service.ServiceError
	var errHandler *HandlerError
	if errors.Is(err, context.DeadlineExceeded) {
		h.logging.Errorf("Method: %s, URL: %v, error: %v", r.Method, r.URL, context.DeadlineExceeded)
		writeResponse(w, r, http.StatusGatewayTimeout, &models.ErrorResponse{
			Code:    http.StatusGatewayTimeout,
			Detail:  "",
			Message: context.DeadlineExceeded.Error(),
		})
	} else if errors.As(err, &errService) {
		h.logging.Errorf("Method: %s, URL: %v, error: %v", r.Method, r.URL, errService.Err)
		writeResponse(w, r, errService.Code, &models.ErrorResponse{
			Code:    errService.Code,
			Detail:  errService.Detail,
			Message: errService.Msg,
		})
	} else if errors.As(err, &errHandler) {
		h.logging.Errorf("Method: %s, URL: %v, error: %v", r.Method, r.URL, errHandler.Err)
		writeResponse(w, r, errHandler.Code, &models.ErrorResponse{
			Code:    errHandler.Code,
			Detail:  errHandler.Detail,
			Message: errHandler.Msg,
		})
	} else {
		e := SystemError(code, err)
		writeResponse(w, r, int64(code), &models.ErrorResponse{
			Code:    e.Code,
			Detail:  e.Detail,
			Message: e.Msg,
		})
		h.logging.Errorf("Method: %s, URL: %v, error: %v", r.Method, r.URL, err)
	}
}

type HandlerError struct {
	// Сама ошибка
	Err error `json:"-"`
	// Код ошибки
	Code int64 `json:"code,omitempty"`
	// Дополнительные детали
	Detail string `json:"detail,omitempty"`
	// Сообщение ошибки
	Msg string `json:"message,omitempty"`
}

func (e *HandlerError) Error() string {
	return e.Msg
}

func (e *HandlerError) Unwrap() error { return e.Err }

func newHandlerError(code int, err error, msg, detail string) *HandlerError {
	return &HandlerError{
		Code:   int64(code),
		Err:    err,
		Msg:    msg,
		Detail: detail,
	}
}

func SystemError(code int, err error) *HandlerError {
	return newHandlerError(code, err, err.Error(), "")
}

func makeValidationError(functionName string, err error) error {
	return newHandlerError(
		code400,
		fmt.Errorf("%s: error: validation failed: %w", functionName, err),
		"validation failed",
		err.Error(),
	)
}

func writeResponse(w http.ResponseWriter, r *http.Request, code int64, resp any) error {
	w.WriteHeader(int(code))

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		return fmt.Errorf("writeResponse/NewEncoder.Encode: %w", err)
	}

	return nil
}
