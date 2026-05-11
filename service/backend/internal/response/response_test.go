package response_test

import (
	"encoding/json"
	"testing"

	"github.com/ldchengyi/linkflow-v2/service/backend/internal/response"
)

func TestSuccessWithoutData(t *testing.T) {
	body := response.Success("ok", 0)

	assertBody(t, body, "ok", 0, nil)
	assertNoDataField(t, body)
}

func TestSuccessWithData(t *testing.T) {
	data := map[string]string{"status": "ok"}
	body := response.SuccessData("ok", 0, data)

	assertBody(t, body, "ok", 0, data)
}

func TestFailWithoutData(t *testing.T) {
	body := response.Fail("bad request", 400)

	assertBody(t, body, "bad request", 400, nil)
	assertNoDataField(t, body)
}

func TestFailWithData(t *testing.T) {
	data := map[string]string{"field": "email"}
	body := response.FailData("invalid input", 10001, data)

	assertBody(t, body, "invalid input", 10001, data)
}

func TestErrorWithoutData(t *testing.T) {
	body := response.Error("internal error", 500)

	assertBody(t, body, "internal error", 500, nil)
	assertNoDataField(t, body)
}

func TestErrorWithData(t *testing.T) {
	data := map[string]string{"trace_id": "trace-1"}
	body := response.ErrorData("internal error", 500, data)

	assertBody(t, body, "internal error", 500, data)
}

func assertBody(t *testing.T, body response.Body, msg string, code int, data any) {
	t.Helper()

	if body.Msg != msg {
		t.Fatalf("Msg = %q, want %q", body.Msg, msg)
	}
	if body.Code != code {
		t.Fatalf("Code = %d, want %d", body.Code, code)
	}
	if data == nil && body.Data != nil {
		t.Fatalf("Data = %#v, want nil", body.Data)
	}
}

func assertNoDataField(t *testing.T, body response.Body) {
	t.Helper()

	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal response body: %v", err)
	}

	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatalf("unmarshal response body: %v", err)
	}
	if _, ok := fields["data"]; ok {
		t.Fatalf("data field exists in %s", raw)
	}
}
