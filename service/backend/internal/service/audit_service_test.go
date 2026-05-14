package service

import (
	"context"
	"errors"
	"testing"
)

type fakeAuditStore struct {
	written AuditEntry
	listed  AuditListInput
	err     error
}

func (f *fakeAuditStore) WriteAuditLog(ctx context.Context, entry AuditEntry) error {
	f.written = entry
	return f.err
}

func (f *fakeAuditStore) ListAuditLogs(ctx context.Context, in AuditListInput) (PageResult[AuditEntry], error) {
	f.listed = in
	if f.err != nil {
		return PageResult[AuditEntry]{}, f.err
	}
	return NewPageResult([]AuditEntry{{ID: "audit-1"}}, 1, in.PageInput), nil
}

func TestAuditServiceListNormalizesPagination(t *testing.T) {
	store := &fakeAuditStore{}
	svc := newTestAuditService(t, store)

	result, err := svc.List(context.Background(), AuditListInput{
		UserID: " user-1 ",
		PageInput: PageInput{
			Page:     2,
			PageSize: 500,
		},
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if store.listed.UserID != "user-1" {
		t.Fatalf("UserID = %q, want user-1", store.listed.UserID)
	}
	if store.listed.Page != 2 {
		t.Fatalf("Page = %d, want 2", store.listed.Page)
	}
	if store.listed.PageSize != maxPageSize {
		t.Fatalf("PageSize = %d, want %d", store.listed.PageSize, maxPageSize)
	}
	if result.PageSize != maxPageSize {
		t.Fatalf("result PageSize = %d, want %d", result.PageSize, maxPageSize)
	}
}

func TestAuditServiceListValidatesUser(t *testing.T) {
	svc := newTestAuditService(t, &fakeAuditStore{})

	if _, err := svc.List(context.Background(), AuditListInput{}); !errors.Is(err, ErrInvalidAuditInput) {
		t.Fatalf("List() error = %v, want ErrInvalidAuditInput", err)
	}
}

func newTestAuditService(t *testing.T, store AuditStore) *AuditService {
	t.Helper()

	svc, err := NewAuditService(store)
	if err != nil {
		t.Fatalf("NewAuditService() error = %v", err)
	}
	return svc
}
