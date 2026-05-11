package httperror_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/ldchengyi/linkflow-v2/service/backend/internal/httperror"
)

func TestFromKeepsHTTPError(t *testing.T) {
	err := httperror.Wrap(httperror.ErrInvalidRequest, errors.New("email is required"))

	got := httperror.From(err)

	if got.HTTPStatus != http.StatusBadRequest {
		t.Fatalf("HTTPStatus = %d, want %d", got.HTTPStatus, http.StatusBadRequest)
	}
	if got.Code != "invalid_request" {
		t.Fatalf("Code = %q, want invalid_request", got.Code)
	}
}

func TestFromWrapsUnknownError(t *testing.T) {
	got := httperror.From(errors.New("db failed"))

	if got.HTTPStatus != http.StatusInternalServerError {
		t.Fatalf("HTTPStatus = %d, want %d", got.HTTPStatus, http.StatusInternalServerError)
	}
	if got.Code != "internal_error" {
		t.Fatalf("Code = %q, want internal_error", got.Code)
	}
}
