package service

import (
	"context"
	"errors"
	"testing"
)

type fakeTenantStore struct {
	created TenantCreateInput
	listed  TenantListInput
	got     TenantGetInput
	updated TenantUpdateInput
	deleted TenantDeleteInput
	tenant  Tenant
	err     error
}

func (f *fakeTenantStore) CreateTenant(ctx context.Context, in TenantCreateInput) (Tenant, error) {
	f.created = in
	if f.err != nil {
		return Tenant{}, f.err
	}
	f.tenant = Tenant{
		ID:          "tenant-1",
		OwnerUserID: in.UserID,
		TenantSlug:  in.TenantSlug,
		TenantName:  in.TenantName,
		Status:      activeTenantStatus,
	}
	return f.tenant, nil
}

func (f *fakeTenantStore) ListTenants(ctx context.Context, in TenantListInput) (PageResult[Tenant], error) {
	f.listed = in
	if f.err != nil {
		return PageResult[Tenant]{}, f.err
	}
	return NewPageResult([]Tenant{f.tenant}, 1, in.PageInput), nil
}

func (f *fakeTenantStore) FindTenantByID(ctx context.Context, in TenantGetInput) (Tenant, error) {
	f.got = in
	if f.err != nil {
		return Tenant{}, f.err
	}
	return f.tenant, nil
}

func (f *fakeTenantStore) UpdateTenant(ctx context.Context, in TenantUpdateInput) (Tenant, error) {
	f.updated = in
	if f.err != nil {
		return Tenant{}, f.err
	}
	f.tenant.TenantName = in.TenantName
	f.tenant.Status = in.Status
	return f.tenant, nil
}

func (f *fakeTenantStore) DeleteTenant(ctx context.Context, in TenantDeleteInput) error {
	f.deleted = in
	return f.err
}

func TestTenantServiceCreateNormalizesSlug(t *testing.T) {
	store := &fakeTenantStore{}
	svc := newTestTenantService(t, store)

	tenant, err := svc.Create(context.Background(), TenantCreateInput{
		UserID:     "user-1",
		TenantSlug: " My-Lab ",
		TenantName: "  My Lab  ",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if store.created.TenantSlug != "my-lab" {
		t.Fatalf("TenantSlug = %q, want my-lab", store.created.TenantSlug)
	}
	if store.created.TenantName != "My Lab" {
		t.Fatalf("TenantName = %q, want My Lab", store.created.TenantName)
	}
	if tenant.OwnerUserID != "user-1" {
		t.Fatalf("OwnerUserID = %q, want user-1", tenant.OwnerUserID)
	}
}

func TestTenantServiceCreateValidatesInput(t *testing.T) {
	svc := newTestTenantService(t, &fakeTenantStore{})

	if _, err := svc.Create(context.Background(), TenantCreateInput{
		TenantSlug: "tenant-1",
		TenantName: "Tenant",
	}); !errors.Is(err, ErrInvalidTenantInput) {
		t.Fatalf("Create() error = %v, want ErrInvalidTenantInput", err)
	}
	if _, err := svc.Create(context.Background(), TenantCreateInput{
		UserID:     "user-1",
		TenantName: "Tenant",
	}); !errors.Is(err, ErrInvalidTenantInput) {
		t.Fatalf("Create() error = %v, want ErrInvalidTenantInput", err)
	}
}

func TestTenantServiceListValidatesUser(t *testing.T) {
	svc := newTestTenantService(t, &fakeTenantStore{})

	if _, err := svc.List(context.Background(), TenantListInput{}); !errors.Is(err, ErrInvalidTenantInput) {
		t.Fatalf("List() error = %v, want ErrInvalidTenantInput", err)
	}
}

func TestTenantServiceGetPassesInput(t *testing.T) {
	store := &fakeTenantStore{tenant: Tenant{ID: "tenant-1"}}
	svc := newTestTenantService(t, store)

	if _, err := svc.Get(context.Background(), TenantGetInput{
		UserID:   "user-1",
		TenantID: "tenant-1",
	}); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if store.got.UserID != "user-1" || store.got.TenantID != "tenant-1" {
		t.Fatalf("Get input = %+v, want user-1/tenant-1", store.got)
	}
}

func TestTenantServiceUpdateDefaultsStatus(t *testing.T) {
	store := &fakeTenantStore{}
	svc := newTestTenantService(t, store)

	if _, err := svc.Update(context.Background(), TenantUpdateInput{
		UserID:     "user-1",
		TenantID:   "tenant-1",
		TenantName: "Tenant",
	}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if store.updated.Status != activeTenantStatus {
		t.Fatalf("Status = %q, want %q", store.updated.Status, activeTenantStatus)
	}
}

func TestTenantServiceUpdateRejectsInvalidStatus(t *testing.T) {
	svc := newTestTenantService(t, &fakeTenantStore{})

	if _, err := svc.Update(context.Background(), TenantUpdateInput{
		UserID:     "user-1",
		TenantID:   "tenant-1",
		TenantName: "Tenant",
		Status:     "deleted",
	}); !errors.Is(err, ErrInvalidTenantInput) {
		t.Fatalf("Update() error = %v, want ErrInvalidTenantInput", err)
	}
}

func TestTenantServiceDeletePassesInput(t *testing.T) {
	store := &fakeTenantStore{}
	svc := newTestTenantService(t, store)

	if err := svc.Delete(context.Background(), TenantDeleteInput{
		UserID:   "user-1",
		TenantID: "tenant-1",
	}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if store.deleted.UserID != "user-1" || store.deleted.TenantID != "tenant-1" {
		t.Fatalf("Delete input = %+v, want user-1/tenant-1", store.deleted)
	}
}

func newTestTenantService(t *testing.T, store TenantStore) *TenantService {
	t.Helper()

	svc, err := NewTenantService(store)
	if err != nil {
		t.Fatalf("NewTenantService() error = %v", err)
	}
	return svc
}
