package service

import (
	"net/http"
)

const (
	code500 = http.StatusInternalServerError
	code400 = http.StatusBadRequest
	code401 = http.StatusUnauthorized

	InternalError       = "internal error"
	UserNotExist        = "user doesn't exist"
	AdminNotExist        = "admin doesn't exist"
	RoleNotExist 		= "a role doesn't exist"
	AuthMessageNotExist = "auth message doesn't exist"
	ParseTokenFailed    = "parse token failed"

	TokenWrongSecret    = "wrong token secret"
	AuthMessageExpired  = "auth message expired"
	WrongSignature      = "wrong signature"
	EcrecoverFailed     = "ecrecover failed"
	CreateUserFailed    = "create user failed"
	UserWithEmailExists = "user with email exists"
	IncorrectPassword   = "incorrect password"
	InvalidAccessParam  = "invalid access parametr"
	UserIsBlocked       = "user is blocked"
	NoEthTxLogs         = "no transaction logs"
	LoginExist          = "login exist"
)

// error struct
type ServiceError struct {
	// Сама ошибка
	Err error `json:"-"`
	// Код ошибки
	Code int64 `json:"code,omitempty"`
	// Дополнительные детали
	Detail string `json:"detail,omitempty"`
	// Сообщение ошибки
	Msg string `json:"message,omitempty"`
}

func (e *ServiceError) Error() string {
	return e.Err.Error()
}

func (e *ServiceError) Unwrap() error { return e.Err }

func newServiceError(code int, err error, msg, detail string) *ServiceError {
	return &ServiceError{
		Code:   int64(code),
		Err:    err,
		Msg:    msg,
		Detail: detail,
	}
}