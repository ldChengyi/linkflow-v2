package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidTenantInput  = errors.New("invalid tenant input")
	ErrTenantAlreadyExists = errors.New("tenant already exists")
	ErrTenantNotFound      = errors.New("tenant not found")
)

const (
	activeTenantStatus   = "active"
	disabledTenantStatus = "disabled"
)

type Tenant struct {
	ID          string    `json:"id"`
	OwnerUserID string    `json:"owner_user_id"`
	TenantSlug  string    `json:"tenant_slug"`
	TenantName  string    `json:"tenant_name"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TenantCreateInput struct {
	UserID     string
	TenantSlug string
	TenantName string
}

type TenantUpdateInput struct {
	UserID     string
	TenantID   string
	TenantName string
	Status     string
}

type TenantGetInput struct {
	UserID   string
	TenantID string
}

type TenantDeleteInput struct {
	UserID   string
	TenantID string
}

type TenantListInput struct {
	UserID string
	PageInput
}

type TenantStore interface {
	CreateTenant(ctx context.Context, in TenantCreateInput) (Tenant, error)
	ListTenants(ctx context.Context, in TenantListInput) (PageResult[Tenant], error)
	FindTenantByID(ctx context.Context, in TenantGetInput) (Tenant, error)
	UpdateTenant(ctx context.Context, in TenantUpdateInput) (Tenant, error)
	DeleteTenant(ctx context.Context, in TenantDeleteInput) error
}

type TenantService struct {
	tenants TenantStore
}

func NewTenantService(tenants TenantStore) (*TenantService, error) {
	if tenants == nil {
		return nil, fmt.Errorf("tenant store is nil")
	}
	return &TenantService{tenants: tenants}, nil
}

func (s *TenantService) Create(ctx context.Context, in TenantCreateInput) (Tenant, error) {
	if err := ctx.Err(); err != nil {
		return Tenant{}, err
	}
	in.UserID = strings.TrimSpace(in.UserID)
	in.TenantSlug = normalizeTenantSlug(in.TenantSlug)
	in.TenantName = strings.TrimSpace(in.TenantName)
	if in.UserID == "" || in.TenantSlug == "" || in.TenantName == "" {
		return Tenant{}, ErrInvalidTenantInput
	}
	return s.tenants.CreateTenant(ctx, in)
}

func (s *TenantService) List(ctx context.Context, in TenantListInput) (PageResult[Tenant], error) {
	if err := ctx.Err(); err != nil {
		return PageResult[Tenant]{}, err
	}
	in.UserID = strings.TrimSpace(in.UserID)
	if in.UserID == "" {
		return PageResult[Tenant]{}, ErrInvalidTenantInput
	}
	in.PageInput = NormalizePageInput(in.PageInput)
	return s.tenants.ListTenants(ctx, in)
}

func (s *TenantService) Get(ctx context.Context, in TenantGetInput) (Tenant, error) {
	if err := ctx.Err(); err != nil {
		return Tenant{}, err
	}
	in.UserID = strings.TrimSpace(in.UserID)
	in.TenantID = strings.TrimSpace(in.TenantID)
	if in.UserID == "" || in.TenantID == "" {
		return Tenant{}, ErrInvalidTenantInput
	}
	return s.tenants.FindTenantByID(ctx, in)
}

func (s *TenantService) Update(ctx context.Context, in TenantUpdateInput) (Tenant, error) {
	if err := ctx.Err(); err != nil {
		return Tenant{}, err
	}
	in.UserID = strings.TrimSpace(in.UserID)
	in.TenantID = strings.TrimSpace(in.TenantID)
	in.TenantName = strings.TrimSpace(in.TenantName)
	in.Status = strings.TrimSpace(in.Status)
	if in.UserID == "" || in.TenantID == "" || in.TenantName == "" {
		return Tenant{}, ErrInvalidTenantInput
	}
	if in.Status == "" {
		in.Status = activeTenantStatus
	}
	if in.Status != activeTenantStatus && in.Status != disabledTenantStatus {
		return Tenant{}, ErrInvalidTenantInput
	}
	return s.tenants.UpdateTenant(ctx, in)
}

func (s *TenantService) Delete(ctx context.Context, in TenantDeleteInput) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	in.UserID = strings.TrimSpace(in.UserID)
	in.TenantID = strings.TrimSpace(in.TenantID)
	if in.UserID == "" || in.TenantID == "" {
		return ErrInvalidTenantInput
	}
	return s.tenants.DeleteTenant(ctx, in)
}

func normalizeTenantSlug(slug string) string {
	return strings.ToLower(strings.TrimSpace(slug))
}
