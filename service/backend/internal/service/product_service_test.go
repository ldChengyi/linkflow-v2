package service

import (
	"context"
	"errors"
	"testing"
)

type fakeProductStore struct {
	created ProductCreateInput
	listed  ProductListInput
	got     ProductGetInput
	updated ProductUpdateInput
	deleted ProductDeleteInput
	product Product
	err     error
}

func (f *fakeProductStore) CreateProduct(ctx context.Context, in ProductCreateInput) (Product, error) {
	f.created = in
	if f.err != nil {
		return Product{}, f.err
	}
	f.product = Product{
		ID:           "product-1",
		TenantID:     in.TenantID,
		ProductKey:   in.ProductKey,
		ProductName:  in.ProductName,
		Description:  in.Description,
		NodeType:     in.NodeType,
		AuthType:     in.AuthType,
		ProtocolType: in.ProtocolType,
		Status:       activeProductStatus,
	}
	return f.product, nil
}

func (f *fakeProductStore) ListProducts(ctx context.Context, in ProductListInput) (PageResult[Product], error) {
	f.listed = in
	if f.err != nil {
		return PageResult[Product]{}, f.err
	}
	return NewPageResult([]Product{f.product}, 1, in.PageInput), nil
}

func (f *fakeProductStore) FindProductByID(ctx context.Context, in ProductGetInput) (Product, error) {
	f.got = in
	if f.err != nil {
		return Product{}, f.err
	}
	return f.product, nil
}

func (f *fakeProductStore) UpdateProduct(ctx context.Context, in ProductUpdateInput) (Product, error) {
	f.updated = in
	if f.err != nil {
		return Product{}, f.err
	}
	f.product.ProductName = in.ProductName
	f.product.Description = in.Description
	f.product.NodeType = in.NodeType
	f.product.AuthType = in.AuthType
	f.product.ProtocolType = in.ProtocolType
	f.product.Status = in.Status
	return f.product, nil
}

func (f *fakeProductStore) DeleteProduct(ctx context.Context, in ProductDeleteInput) error {
	f.deleted = in
	return f.err
}

func TestProductServiceCreateNormalizesInputAndDefaults(t *testing.T) {
	store := &fakeProductStore{}
	svc := newTestProductService(t, store)

	product, err := svc.Create(context.Background(), ProductCreateInput{
		UserID:      "user-1",
		TenantID:    "tenant-1",
		ProductKey:  " ESP32-Thermo ",
		ProductName: "  ESP32 Thermo  ",
		Description: "  lab sensor  ",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if store.created.ProductKey != "esp32-thermo" {
		t.Fatalf("ProductKey = %q, want esp32-thermo", store.created.ProductKey)
	}
	if store.created.ProductName != "ESP32 Thermo" {
		t.Fatalf("ProductName = %q, want ESP32 Thermo", store.created.ProductName)
	}
	if store.created.Description != "lab sensor" {
		t.Fatalf("Description = %q, want lab sensor", store.created.Description)
	}
	if store.created.NodeType != defaultProductNodeType || store.created.AuthType != defaultProductAuthType || store.created.ProtocolType != defaultProductProtocolType {
		t.Fatalf("defaults = %q/%q/%q, want direct/secret/mqtt", store.created.NodeType, store.created.AuthType, store.created.ProtocolType)
	}
	if product.TenantID != "tenant-1" {
		t.Fatalf("TenantID = %q, want tenant-1", product.TenantID)
	}
}

func TestProductServiceCreateValidatesRequiredInput(t *testing.T) {
	svc := newTestProductService(t, &fakeProductStore{})

	if _, err := svc.Create(context.Background(), ProductCreateInput{
		TenantID:    "tenant-1",
		ProductKey:  "esp32",
		ProductName: "ESP32",
	}); !errors.Is(err, ErrInvalidProductInput) {
		t.Fatalf("Create() error = %v, want ErrInvalidProductInput", err)
	}
	if _, err := svc.Create(context.Background(), ProductCreateInput{
		UserID:      "user-1",
		ProductKey:  "esp32",
		ProductName: "ESP32",
	}); !errors.Is(err, ErrInvalidProductInput) {
		t.Fatalf("Create() error = %v, want ErrInvalidProductInput", err)
	}
}

func TestProductServiceCreateRejectsInvalidVariants(t *testing.T) {
	svc := newTestProductService(t, &fakeProductStore{})

	if _, err := svc.Create(context.Background(), ProductCreateInput{
		UserID:      "user-1",
		TenantID:    "tenant-1",
		ProductKey:  "esp32",
		ProductName: "ESP32",
		NodeType:    "unknown",
	}); !errors.Is(err, ErrInvalidProductInput) {
		t.Fatalf("Create() error = %v, want ErrInvalidProductInput", err)
	}
	if _, err := svc.Create(context.Background(), ProductCreateInput{
		UserID:      "user-1",
		TenantID:    "tenant-1",
		ProductKey:  "esp32",
		ProductName: "ESP32",
		AuthType:    "token",
	}); !errors.Is(err, ErrInvalidProductInput) {
		t.Fatalf("Create() error = %v, want ErrInvalidProductInput", err)
	}
	if _, err := svc.Create(context.Background(), ProductCreateInput{
		UserID:       "user-1",
		TenantID:     "tenant-1",
		ProductKey:   "esp32",
		ProductName:  "ESP32",
		ProtocolType: "ftp",
	}); !errors.Is(err, ErrInvalidProductInput) {
		t.Fatalf("Create() error = %v, want ErrInvalidProductInput", err)
	}
}

func TestProductServiceListNormalizesPagination(t *testing.T) {
	store := &fakeProductStore{}
	svc := newTestProductService(t, store)

	if _, err := svc.List(context.Background(), ProductListInput{
		UserID:    "user-1",
		TenantID:  "tenant-1",
		PageInput: PageInput{Page: -1, PageSize: 1000},
	}); err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if store.listed.Page != defaultPage || store.listed.PageSize != maxPageSize {
		t.Fatalf("PageInput = %+v, want page %d page_size %d", store.listed.PageInput, defaultPage, maxPageSize)
	}
}

func TestProductServiceUpdateDefaultsAndValidatesStatus(t *testing.T) {
	store := &fakeProductStore{}
	svc := newTestProductService(t, store)

	if _, err := svc.Update(context.Background(), ProductUpdateInput{
		UserID:      "user-1",
		ProductID:   "product-1",
		ProductName: "ESP32",
	}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if store.updated.NodeType != defaultProductNodeType || store.updated.AuthType != defaultProductAuthType || store.updated.ProtocolType != defaultProductProtocolType || store.updated.Status != activeProductStatus {
		t.Fatalf("defaults = %+v, want direct/secret/mqtt/active", store.updated)
	}

	if _, err := svc.Update(context.Background(), ProductUpdateInput{
		UserID:      "user-1",
		ProductID:   "product-1",
		ProductName: "ESP32",
		Status:      "deleted",
	}); !errors.Is(err, ErrInvalidProductInput) {
		t.Fatalf("Update() error = %v, want ErrInvalidProductInput", err)
	}
}

func TestProductServiceGetAndDeletePassInput(t *testing.T) {
	store := &fakeProductStore{product: Product{ID: "product-1"}}
	svc := newTestProductService(t, store)

	if _, err := svc.Get(context.Background(), ProductGetInput{
		UserID:    "user-1",
		ProductID: "product-1",
	}); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if store.got.UserID != "user-1" || store.got.ProductID != "product-1" {
		t.Fatalf("Get input = %+v, want user-1/product-1", store.got)
	}

	if err := svc.Delete(context.Background(), ProductDeleteInput{
		UserID:    "user-1",
		ProductID: "product-1",
	}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if store.deleted.UserID != "user-1" || store.deleted.ProductID != "product-1" {
		t.Fatalf("Delete input = %+v, want user-1/product-1", store.deleted)
	}
}

func newTestProductService(t *testing.T, store ProductStore) *ProductService {
	t.Helper()

	svc, err := NewProductService(store)
	if err != nil {
		t.Fatalf("NewProductService() error = %v", err)
	}
	return svc
}
