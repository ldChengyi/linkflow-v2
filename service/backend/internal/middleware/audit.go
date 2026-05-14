package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/ldchengyi/linkflow-v2/service/backend/internal/service"
)

type auditContextKey struct{}

type AuditRecorder struct {
	mu           sync.Mutex
	action       string
	resourceType string
	resourceID   string
	errorCode    string
	metadata     map[string]any
}

type AuditWriter interface {
	Write(ctx context.Context, entry service.AuditEntry) error
}

func Audit(writer AuditWriter, log *slog.Logger) Middleware {
	if log == nil {
		log = slog.Default()
	}
	return func(next http.Handler) http.Handler {
		if writer == nil {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			recorder := &AuditRecorder{metadata: map[string]any{}}
			ctx := context.WithValue(r.Context(), auditContextKey{}, recorder)
			r = r.WithContext(ctx)

			rw := &auditResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			start := time.Now()
			next.ServeHTTP(rw, r)

			entry, ok := recorder.entry()
			if !ok {
				return
			}

			principal, _ := PrincipalFromContext(r.Context())
			entry.ActorUserID = principal.UserID
			entry.ActorRole = principal.Role
			entry.Method = r.Method
			entry.Path = r.URL.Path
			entry.StatusCode = rw.statusCode
			entry.DurationMS = int(time.Since(start).Milliseconds())
			entry.IP = clientIP(r)
			entry.UserAgent = r.UserAgent()
			entry.RequestID = r.Header.Get("X-Request-Id")
			entry.TraceID = r.Header.Get("Traceparent")
			entry.OperationID = r.Header.Get("X-Operation-Id")
			if rw.statusCode >= http.StatusBadRequest {
				entry.Result = service.AuditResultFailure
				if entry.ErrorCode == "" {
					entry.ErrorCode = strconv.Itoa(rw.statusCode)
				}
			} else {
				entry.Result = service.AuditResultSuccess
			}

			if err := writer.Write(context.WithoutCancel(r.Context()), entry); err != nil {
				log.Error("write audit log", "err", err, "action", entry.Action, "resource_type", entry.ResourceType)
			}
		})
	}
}

func AuditFromContext(ctx context.Context) (*AuditRecorder, bool) {
	recorder, ok := ctx.Value(auditContextKey{}).(*AuditRecorder)
	return recorder, ok
}

func (r *AuditRecorder) Set(action string, resourceType string, resourceID string, metadata map[string]any) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.action = action
	r.resourceType = resourceType
	r.resourceID = resourceID
	if metadata != nil {
		if r.metadata == nil {
			r.metadata = map[string]any{}
		}
		for k, v := range metadata {
			r.metadata[k] = v
		}
	}
}

func (r *AuditRecorder) SetResourceID(resourceID string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.resourceID = resourceID
}

func (r *AuditRecorder) SetErrorCode(code string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.errorCode = code
}

func (r *AuditRecorder) AddMetadata(metadata map[string]any) {
	if r == nil || metadata == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.metadata == nil {
		r.metadata = map[string]any{}
	}
	for k, v := range metadata {
		r.metadata[k] = v
	}
}

func (r *AuditRecorder) entry() (service.AuditEntry, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.action == "" || r.resourceType == "" {
		return service.AuditEntry{}, false
	}
	metadata := make(map[string]any, len(r.metadata))
	for k, v := range r.metadata {
		metadata[k] = v
	}
	return service.AuditEntry{
		Action:       r.action,
		ResourceType: r.resourceType,
		ResourceID:   r.resourceID,
		ErrorCode:    r.errorCode,
		Metadata:     metadata,
	}, true
}

type auditResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *auditResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}
	if realIP := r.Header.Get("X-Real-Ip"); realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		if errors.Is(err, net.ErrClosed) {
			return ""
		}
		return r.RemoteAddr
	}
	return host
}
