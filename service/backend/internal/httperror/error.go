package httperror

import (
	"errors"
	"net/http"
)

var (
	ErrBadRequest          = New(http.StatusBadRequest, "bad_request", "bad request")
	ErrUnauthorized        = New(http.StatusUnauthorized, "unauthorized", "unauthorized")
	ErrForbidden           = New(http.StatusForbidden, "forbidden", "forbidden")
	ErrNotFound            = New(http.StatusNotFound, "not_found", "not found")
	ErrConflict            = New(http.StatusConflict, "conflict", "conflict")
	ErrUnsupportedMedia    = New(http.StatusUnsupportedMediaType, "unsupported_media_type", "unsupported media type")
	ErrUnprocessableEntity = New(http.StatusUnprocessableEntity, "unprocessable_entity", "unprocessable entity")
	ErrInternal            = New(http.StatusInternalServerError, "internal_error", "internal server error")
	ErrServiceUnavailable  = New(http.StatusServiceUnavailable, "service_unavailable", "service unavailable")

	ErrInvalidJSON         = New(http.StatusBadRequest, "invalid_json", "invalid json body")
	ErrInvalidRequest      = New(http.StatusBadRequest, "invalid_request", "invalid request")
	ErrInvalidCredentials  = New(http.StatusUnauthorized, "invalid_credentials", "invalid credentials")
	ErrUserAlreadyExists   = New(http.StatusConflict, "user_already_exists", "user already exists")
	ErrUserDisabled        = New(http.StatusForbidden, "user_disabled", "user is disabled")
	ErrTenantAlreadyExists = New(http.StatusConflict, "tenant_already_exists", "tenant already exists")
	ErrMethodNotAllowed    = New(http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
	ErrContentTypeRequired = New(http.StatusUnsupportedMediaType, "content_type_required", "content type must be application/json")
)

type Error struct {
	HTTPStatus int
	Code       string
	Msg        string
	Err        error
}

func New(status int, code string, msg string) *Error {
	return &Error{
		HTTPStatus: status,
		Code:       code,
		Msg:        msg,
	}
}

func Wrap(base *Error, err error) *Error {
	if base == nil {
		base = ErrInternal
	}
	return &Error{
		HTTPStatus: base.HTTPStatus,
		Code:       base.Code,
		Msg:        base.Msg,
		Err:        err,
	}
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil {
		return e.Msg + ": " + e.Err.Error()
	}
	return e.Msg
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func From(err error) *Error {
	if err == nil {
		return nil
	}

	var httpErr *Error
	if errors.As(err, &httpErr) {
		return httpErr
	}

	return Wrap(ErrInternal, err)
}
