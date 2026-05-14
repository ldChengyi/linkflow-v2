package middleware

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ldchengyi/linkflow-v2/service/backend/internal/domain"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/service"
)

type fakeAuditWriter struct {
	entries []service.AuditEntry
}

func (f *fakeAuditWriter) Write(ctx context.Context, entry service.AuditEntry) error {
	f.entries = append(f.entries, entry)
	return nil
}

func TestAuditSkipsUnmarkedRequest(t *testing.T) {
	writer := &fakeAuditWriter{}
	handler := Audit(writer, slog.New(slog.NewTextHandler(io.Discard, nil)))(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tenants", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if len(writer.entries) != 0 {
		t.Fatalf("audit entries = %d, want 0", len(writer.entries))
	}
}

func TestAuditWritesMarkedRequest(t *testing.T) {
	writer := &fakeAuditWriter{}
	handler := Audit(writer, slog.New(slog.NewTextHandler(io.Discard, nil)))(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			recorder, ok := AuditFromContext(r.Context())
			if !ok {
				t.Fatal("audit recorder missing from context")
			}
			recorder.Set("tenant.update", "tenant", "tenant-1", map[string]any{
				"tenant_slug": "my-lab",
			})
			w.WriteHeader(http.StatusCreated)
		}),
	)

	ctx := ContextWithPrincipal(context.Background(), domain.Principal{
		UserID: "user-1",
		Role:   "user",
	})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/tenants/tenant-1", nil).WithContext(ctx)
	req.Header.Set("User-Agent", "audit-test")
	req.Header.Set("X-Request-Id", "req-1")
	req.Header.Set("X-Operation-Id", "op-1")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if len(writer.entries) != 1 {
		t.Fatalf("audit entries = %d, want 1", len(writer.entries))
	}
	entry := writer.entries[0]
	if entry.Action != "tenant.update" || entry.ResourceType != "tenant" || entry.ResourceID != "tenant-1" {
		t.Fatalf("audit resource = %s/%s/%s, want tenant.update/tenant/tenant-1", entry.Action, entry.ResourceType, entry.ResourceID)
	}
	if entry.ActorUserID != "user-1" || entry.ActorRole != "user" {
		t.Fatalf("actor = %s/%s, want user-1/user", entry.ActorUserID, entry.ActorRole)
	}
	if entry.Method != http.MethodPut || entry.Path != "/api/v1/tenants/tenant-1" || entry.StatusCode != http.StatusCreated {
		t.Fatalf("http audit = %s %s %d, want PUT /api/v1/tenants/tenant-1 201", entry.Method, entry.Path, entry.StatusCode)
	}
	if entry.Result != service.AuditResultSuccess {
		t.Fatalf("result = %q, want success", entry.Result)
	}
	if entry.RequestID != "req-1" || entry.OperationID != "op-1" {
		t.Fatalf("ids = %s/%s, want req-1/op-1", entry.RequestID, entry.OperationID)
	}
	if entry.Metadata["tenant_slug"] != "my-lab" {
		t.Fatalf("metadata tenant_slug = %v, want my-lab", entry.Metadata["tenant_slug"])
	}
}
