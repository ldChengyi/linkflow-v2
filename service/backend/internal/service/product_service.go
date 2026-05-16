package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidProductInput  = errors.New("invalid product input")
	ErrProductAlreadyExists = errors.New("product already exists")
	ErrProductNotFound      = errors.New("product not found")
)

const (
	defaultProductNodeType     = "direct"
	gatewayProductNodeType     = "gateway"
	subDeviceProductNodeType   = "sub_device"
	defaultProductAuthType     = "secret"
	certificateProductAuthType = "certificate"
	anonymousProductAuthType   = "anonymous"
	defaultProductProtocolType = "mqtt"
	httpProductProtocolType    = "http"
	coapProductProtocolType    = "coap"
	modbusProductProtocolType  = "modbus"
	opcuaProductProtocolType   = "opcua"
	loraProductProtocolType    = "lora"
	activeProductStatus        = "active"
	disabledProductStatus      = "disabled"
)

type Product struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	ProductKey   string    `json:"product_key"`
	ProductName  string    `json:"product_name"`
	Description  string    `json:"description"`
	NodeType     string    `json:"node_type"`
	AuthType     string    `json:"auth_type"`
	ProtocolType string    `json:"protocol_type"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ProductCreateInput struct {
	UserID       string
	TenantID     string
	ProductKey   string
	ProductName  string
	Description  string
	NodeType     string
	AuthType     string
	ProtocolType string
}

type ProductUpdateInput struct {
	UserID       string
	ProductID    string
	ProductName  string
	Description  string
	NodeType     string
	AuthType     string
	ProtocolType string
	Status       string
}

type ProductGetInput struct {
	UserID    string
	ProductID string
}

type ProductDeleteInput struct {
	UserID    string
	ProductID string
}

type ProductListInput struct {
	UserID   string
	TenantID string
	PageInput
}

type ProductStore interface {
	CreateProduct(ctx context.Context, in ProductCreateInput) (Product, error)
	ListProducts(ctx context.Context, in ProductListInput) (PageResult[Product], error)
	FindProductByID(ctx context.Context, in ProductGetInput) (Product, error)
	UpdateProduct(ctx context.Context, in ProductUpdateInput) (Product, error)
	DeleteProduct(ctx context.Context, in ProductDeleteInput) error
}

type ProductService struct {
	products ProductStore
}

func NewProductService(products ProductStore) (*ProductService, error) {
	if products == nil {
		return nil, fmt.Errorf("product store is nil")
	}
	return &ProductService{products: products}, nil
}

func (s *ProductService) Create(ctx context.Context, in ProductCreateInput) (Product, error) {
	if err := ctx.Err(); err != nil {
		return Product{}, err
	}
	in.UserID = strings.TrimSpace(in.UserID)
	in.TenantID = strings.TrimSpace(in.TenantID)
	in.ProductKey = normalizeProductKey(in.ProductKey)
	in.ProductName = strings.TrimSpace(in.ProductName)
	in.Description = strings.TrimSpace(in.Description)
	in.NodeType = normalizeProductNodeType(in.NodeType)
	in.AuthType = normalizeProductAuthType(in.AuthType)
	in.ProtocolType = normalizeProductProtocolType(in.ProtocolType)
	if in.UserID == "" || in.TenantID == "" || in.ProductKey == "" || in.ProductName == "" {
		return Product{}, ErrInvalidProductInput
	}
	if !validProductNodeType(in.NodeType) || !validProductAuthType(in.AuthType) || !validProductProtocolType(in.ProtocolType) {
		return Product{}, ErrInvalidProductInput
	}
	return s.products.CreateProduct(ctx, in)
}

func (s *ProductService) List(ctx context.Context, in ProductListInput) (PageResult[Product], error) {
	if err := ctx.Err(); err != nil {
		return PageResult[Product]{}, err
	}
	in.UserID = strings.TrimSpace(in.UserID)
	in.TenantID = strings.TrimSpace(in.TenantID)
	if in.UserID == "" || in.TenantID == "" {
		return PageResult[Product]{}, ErrInvalidProductInput
	}
	in.PageInput = NormalizePageInput(in.PageInput)
	return s.products.ListProducts(ctx, in)
}

func (s *ProductService) Get(ctx context.Context, in ProductGetInput) (Product, error) {
	if err := ctx.Err(); err != nil {
		return Product{}, err
	}
	in.UserID = strings.TrimSpace(in.UserID)
	in.ProductID = strings.TrimSpace(in.ProductID)
	if in.UserID == "" || in.ProductID == "" {
		return Product{}, ErrInvalidProductInput
	}
	return s.products.FindProductByID(ctx, in)
}

func (s *ProductService) Update(ctx context.Context, in ProductUpdateInput) (Product, error) {
	if err := ctx.Err(); err != nil {
		return Product{}, err
	}
	in.UserID = strings.TrimSpace(in.UserID)
	in.ProductID = strings.TrimSpace(in.ProductID)
	in.ProductName = strings.TrimSpace(in.ProductName)
	in.Description = strings.TrimSpace(in.Description)
	in.NodeType = normalizeProductNodeType(in.NodeType)
	in.AuthType = normalizeProductAuthType(in.AuthType)
	in.ProtocolType = normalizeProductProtocolType(in.ProtocolType)
	in.Status = normalizeProductStatus(in.Status)
	if in.UserID == "" || in.ProductID == "" || in.ProductName == "" {
		return Product{}, ErrInvalidProductInput
	}
	if !validProductNodeType(in.NodeType) || !validProductAuthType(in.AuthType) || !validProductProtocolType(in.ProtocolType) || !validProductStatus(in.Status) {
		return Product{}, ErrInvalidProductInput
	}
	return s.products.UpdateProduct(ctx, in)
}

func (s *ProductService) Delete(ctx context.Context, in ProductDeleteInput) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	in.UserID = strings.TrimSpace(in.UserID)
	in.ProductID = strings.TrimSpace(in.ProductID)
	if in.UserID == "" || in.ProductID == "" {
		return ErrInvalidProductInput
	}
	return s.products.DeleteProduct(ctx, in)
}

func normalizeProductKey(key string) string {
	return strings.ToLower(strings.TrimSpace(key))
}

func normalizeProductNodeType(nodeType string) string {
	nodeType = strings.TrimSpace(nodeType)
	if nodeType == "" {
		return defaultProductNodeType
	}
	return nodeType
}

func normalizeProductAuthType(authType string) string {
	authType = strings.TrimSpace(authType)
	if authType == "" {
		return defaultProductAuthType
	}
	return authType
}

func normalizeProductProtocolType(protocolType string) string {
	protocolType = strings.TrimSpace(protocolType)
	if protocolType == "" {
		return defaultProductProtocolType
	}
	return protocolType
}

func normalizeProductStatus(status string) string {
	status = strings.TrimSpace(status)
	if status == "" {
		return activeProductStatus
	}
	return status
}

func validProductNodeType(nodeType string) bool {
	switch nodeType {
	case defaultProductNodeType, gatewayProductNodeType, subDeviceProductNodeType:
		return true
	default:
		return false
	}
}

func validProductAuthType(authType string) bool {
	switch authType {
	case defaultProductAuthType, certificateProductAuthType, anonymousProductAuthType:
		return true
	default:
		return false
	}
}

func validProductProtocolType(protocolType string) bool {
	switch protocolType {
	case defaultProductProtocolType, httpProductProtocolType, coapProductProtocolType, modbusProductProtocolType, opcuaProductProtocolType, loraProductProtocolType:
		return true
	default:
		return false
	}
}

func validProductStatus(status string) bool {
	switch status {
	case activeProductStatus, disabledProductStatus:
		return true
	default:
		return false
	}
}
