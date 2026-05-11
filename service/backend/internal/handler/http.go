package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ldchengyi/linkflow-v2/service/backend/internal/httperror"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/response"
)

func decodeJSON(r *http.Request, dst any) error {
	contentType := r.Header.Get("Content-Type")
	if contentType == "" || !strings.HasPrefix(strings.ToLower(contentType), "application/json") {
		return httperror.ErrContentTypeRequired
	}
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return httperror.Wrap(httperror.ErrInvalidJSON, err)
	}

	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return httperror.ErrInvalidJSON
	}

	return nil
}

func writeHTTPError(w http.ResponseWriter, err error) {
	httpErr := httperror.From(err)
	writeJSON(w, httpErr.HTTPStatus, response.Fail(httpErr.Msg, httpErr.HTTPStatus))
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
