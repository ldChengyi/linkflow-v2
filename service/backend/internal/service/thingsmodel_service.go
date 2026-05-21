package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidThingsModelInput  = errors.New("invalid thingsmodel input")
	ErrThingsModelAlreadyExists = errors.New("thingsmodel already exists")
	ErrThingsModelNotFound      = errors.New("thingsmodel not found")
)

const (
	thingsModelStatusDraft      = "draft"
	thingsModelStatusPublished  = "published"
	thingsModelStatusDeprecated = "deprecated"
)

type ThingsModelObject map[string]any

type ThingsModel struct {
	ID           string            `json:"id"`
	TenantID     string            `json:"tenant_id"`
	ProductID    string            `json:"product_id"`
	ModelVersion int               `json:"model_version"`
	ModelName    string            `json:"model_name"`
	Description  string            `json:"description"`
	Status       string            `json:"status"`
	IsCurrent    bool              `json:"is_current"`
	Properties   ThingsModelObject `json:"properties"`
	Events       ThingsModelObject `json:"events"`
	Services     ThingsModelObject `json:"services"`
	PublishedAt  *time.Time        `json:"published_at,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

type ThingsModelCreateInput struct {
	UserID       string
	TenantID     string
	ProductID    string
	ModelVersion int
	ModelName    string
	Description  string
	Status       string
	IsCurrent    bool
	Properties   ThingsModelObject
	Events       ThingsModelObject
	Services     ThingsModelObject
}

type ThingsModelUpdateInput struct {
	UserID        string
	ThingsModelID string
	ModelName     string
	Description   string
	Status        string
	IsCurrent     bool
	Properties    ThingsModelObject
	Events        ThingsModelObject
	Services      ThingsModelObject
}

type ThingsModelGetInput struct {
	UserID        string
	ThingsModelID string
}

type ThingsModelDeleteInput struct {
	UserID        string
	ThingsModelID string
}

type ThingsModelListInput struct {
	UserID    string
	TenantID  string
	ProductID string
	PageInput
}

type ThingsModelStore interface {
	CreateThingsModel(ctx context.Context, in ThingsModelCreateInput) (ThingsModel, error)
	ListThingsModels(ctx context.Context, in ThingsModelListInput) (PageResult[ThingsModel], error)
	FindThingsModelByID(ctx context.Context, in ThingsModelGetInput) (ThingsModel, error)
	UpdateThingsModel(ctx context.Context, in ThingsModelUpdateInput) (ThingsModel, error)
	DeleteThingsModel(ctx context.Context, in ThingsModelDeleteInput) error
}

type ThingsModelService struct {
	models ThingsModelStore
}

func NewThingsModelService(models ThingsModelStore) (*ThingsModelService, error) {
	if models == nil {
		return nil, fmt.Errorf("thingsmodel store is nil")
	}
	return &ThingsModelService{models: models}, nil
}

func (s *ThingsModelService) Create(ctx context.Context, in ThingsModelCreateInput) (ThingsModel, error) {
	if err := ctx.Err(); err != nil {
		return ThingsModel{}, err
	}
	in.UserID = strings.TrimSpace(in.UserID)
	in.TenantID = strings.TrimSpace(in.TenantID)
	in.ProductID = strings.TrimSpace(in.ProductID)
	in.ModelName = strings.TrimSpace(in.ModelName)
	in.Description = strings.TrimSpace(in.Description)
	in.Status = normalizeThingsModelStatus(in.Status)
	in.Properties = normalizeThingsModelObject(in.Properties)
	in.Events = normalizeThingsModelObject(in.Events)
	in.Services = normalizeThingsModelObject(in.Services)
	if in.UserID == "" || in.TenantID == "" || in.ProductID == "" || in.ModelName == "" || in.ModelVersion <= 0 {
		return ThingsModel{}, ErrInvalidThingsModelInput
	}
	if !validThingsModelStatus(in.Status) || (in.IsCurrent && in.Status != thingsModelStatusPublished) {
		return ThingsModel{}, ErrInvalidThingsModelInput
	}
	if err := validateThingsModelDefinition(in.Properties, in.Events, in.Services); err != nil {
		return ThingsModel{}, ErrInvalidThingsModelInput
	}
	return s.models.CreateThingsModel(ctx, in)
}

func (s *ThingsModelService) List(ctx context.Context, in ThingsModelListInput) (PageResult[ThingsModel], error) {
	if err := ctx.Err(); err != nil {
		return PageResult[ThingsModel]{}, err
	}
	in.UserID = strings.TrimSpace(in.UserID)
	in.TenantID = strings.TrimSpace(in.TenantID)
	in.ProductID = strings.TrimSpace(in.ProductID)
	if in.UserID == "" || in.TenantID == "" {
		return PageResult[ThingsModel]{}, ErrInvalidThingsModelInput
	}
	in.PageInput = NormalizePageInput(in.PageInput)
	return s.models.ListThingsModels(ctx, in)
}

func (s *ThingsModelService) Get(ctx context.Context, in ThingsModelGetInput) (ThingsModel, error) {
	if err := ctx.Err(); err != nil {
		return ThingsModel{}, err
	}
	in.UserID = strings.TrimSpace(in.UserID)
	in.ThingsModelID = strings.TrimSpace(in.ThingsModelID)
	if in.UserID == "" || in.ThingsModelID == "" {
		return ThingsModel{}, ErrInvalidThingsModelInput
	}
	return s.models.FindThingsModelByID(ctx, in)
}

func (s *ThingsModelService) Update(ctx context.Context, in ThingsModelUpdateInput) (ThingsModel, error) {
	if err := ctx.Err(); err != nil {
		return ThingsModel{}, err
	}
	in.UserID = strings.TrimSpace(in.UserID)
	in.ThingsModelID = strings.TrimSpace(in.ThingsModelID)
	in.ModelName = strings.TrimSpace(in.ModelName)
	in.Description = strings.TrimSpace(in.Description)
	in.Status = normalizeThingsModelStatus(in.Status)
	in.Properties = normalizeThingsModelObject(in.Properties)
	in.Events = normalizeThingsModelObject(in.Events)
	in.Services = normalizeThingsModelObject(in.Services)
	if in.UserID == "" || in.ThingsModelID == "" || in.ModelName == "" {
		return ThingsModel{}, ErrInvalidThingsModelInput
	}
	if !validThingsModelStatus(in.Status) || (in.IsCurrent && in.Status != thingsModelStatusPublished) {
		return ThingsModel{}, ErrInvalidThingsModelInput
	}
	if err := validateThingsModelDefinition(in.Properties, in.Events, in.Services); err != nil {
		return ThingsModel{}, ErrInvalidThingsModelInput
	}
	return s.models.UpdateThingsModel(ctx, in)
}

func (s *ThingsModelService) Delete(ctx context.Context, in ThingsModelDeleteInput) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	in.UserID = strings.TrimSpace(in.UserID)
	in.ThingsModelID = strings.TrimSpace(in.ThingsModelID)
	if in.UserID == "" || in.ThingsModelID == "" {
		return ErrInvalidThingsModelInput
	}
	return s.models.DeleteThingsModel(ctx, in)
}

func normalizeThingsModelStatus(status string) string {
	status = strings.TrimSpace(status)
	if status == "" {
		return thingsModelStatusDraft
	}
	return status
}

func normalizeThingsModelObject(value ThingsModelObject) ThingsModelObject {
	if value == nil {
		return ThingsModelObject{}
	}
	return value
}

func validThingsModelStatus(status string) bool {
	switch status {
	case thingsModelStatusDraft, thingsModelStatusPublished, thingsModelStatusDeprecated:
		return true
	default:
		return false
	}
}
