package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	AuditResultSuccess = "success"
	AuditResultFailure = "failure"
)

var ErrInvalidAuditInput = errors.New("invalid audit input")

type AuditEntry struct {
	ID           string         `json:"id,omitempty"`
	ActorUserID  string         `json:"actor_user_id"`
	ActorRole    string         `json:"actor_role"`
	Action       string         `json:"action"`
	ResourceType string         `json:"resource_type"`
	ResourceID   string         `json:"resource_id"`
	Result       string         `json:"result"`
	ErrorCode    string         `json:"error_code"`
	Method       string         `json:"method"`
	Path         string         `json:"path"`
	StatusCode   int            `json:"status_code"`
	DurationMS   int            `json:"duration_ms"`
	IP           string         `json:"ip"`
	UserAgent    string         `json:"user_agent"`
	RequestID    string         `json:"request_id"`
	TraceID      string         `json:"trace_id"`
	OperationID  string         `json:"operation_id"`
	Metadata     map[string]any `json:"metadata"`
	CreatedAt    time.Time      `json:"created_at,omitempty"`
}

type AuditListInput struct {
	UserID string
	PageInput
}

type AuditStore interface {
	WriteAuditLog(ctx context.Context, entry AuditEntry) error
	ListAuditLogs(ctx context.Context, in AuditListInput) (PageResult[AuditEntry], error)
}

type AuditService struct {
	store AuditStore
}

func NewAuditService(store AuditStore) (*AuditService, error) {
	if store == nil {
		return nil, fmt.Errorf("audit store is nil")
	}
	return &AuditService{store: store}, nil
}

func (s *AuditService) Write(ctx context.Context, entry AuditEntry) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	entry.ActorUserID = strings.TrimSpace(entry.ActorUserID)
	entry.ActorRole = strings.TrimSpace(entry.ActorRole)
	entry.Action = strings.TrimSpace(entry.Action)
	entry.ResourceType = strings.TrimSpace(entry.ResourceType)
	entry.ResourceID = strings.TrimSpace(entry.ResourceID)
	entry.Result = strings.TrimSpace(entry.Result)
	entry.ErrorCode = strings.TrimSpace(entry.ErrorCode)
	entry.Method = strings.TrimSpace(entry.Method)
	entry.Path = strings.TrimSpace(entry.Path)
	entry.IP = strings.TrimSpace(entry.IP)
	entry.UserAgent = strings.TrimSpace(entry.UserAgent)
	entry.RequestID = strings.TrimSpace(entry.RequestID)
	entry.TraceID = strings.TrimSpace(entry.TraceID)
	entry.OperationID = strings.TrimSpace(entry.OperationID)
	if entry.Action == "" || entry.ResourceType == "" {
		return fmt.Errorf("audit action and resource type are required")
	}
	if entry.Method == "" || entry.Path == "" {
		return fmt.Errorf("audit method and path are required")
	}
	if entry.Result == "" {
		entry.Result = AuditResultSuccess
	}
	if entry.Result != AuditResultSuccess && entry.Result != AuditResultFailure {
		return fmt.Errorf("invalid audit result %q", entry.Result)
	}
	if entry.Metadata == nil {
		entry.Metadata = map[string]any{}
	}
	return s.store.WriteAuditLog(ctx, entry)
}

func (s *AuditService) List(ctx context.Context, in AuditListInput) (PageResult[AuditEntry], error) {
	if err := ctx.Err(); err != nil {
		return PageResult[AuditEntry]{}, err
	}
	in.UserID = strings.TrimSpace(in.UserID)
	if in.UserID == "" {
		return PageResult[AuditEntry]{}, ErrInvalidAuditInput
	}
	in.PageInput = NormalizePageInput(in.PageInput)
	return s.store.ListAuditLogs(ctx, in)
}

func AuditMetadataJSON(metadata map[string]any) ([]byte, error) {
	if metadata == nil {
		metadata = map[string]any{}
	}
	raw, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("marshal audit metadata: %w", err)
	}
	return raw, nil
}
