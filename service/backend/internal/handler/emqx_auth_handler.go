package handler

import (
	"crypto/subtle"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/ldchengyi/linkflow-v2/service/backend/internal/httperror"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/service"
)

type EMQXAuthHandler struct {
	service         *service.MQTTAuthService
	serviceUsername string
	servicePassword string
	log             *slog.Logger
}

type emqxAuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	ClientID string `json:"clientid"`
}

type emqxAuthResponse struct {
	Result      string            `json:"result"`
	IsSuperuser bool              `json:"is_superuser,omitempty"`
	ClientAttrs map[string]string `json:"client_attrs,omitempty"`
}

func NewEMQXAuthHandler(service *service.MQTTAuthService, serviceUsername string, servicePassword string, log *slog.Logger) (*EMQXAuthHandler, error) {
	if service == nil {
		return nil, errors.New("mqtt auth service is nil")
	}
	serviceUsername = strings.TrimSpace(serviceUsername)
	servicePassword = strings.TrimSpace(servicePassword)
	if serviceUsername == "" || servicePassword == "" {
		return nil, errors.New("emqx service client credentials are required")
	}
	if log == nil {
		log = slog.Default()
	}
	return &EMQXAuthHandler{service: service, serviceUsername: serviceUsername, servicePassword: servicePassword, log: log}, nil
}

func (h *EMQXAuthHandler) RegisterRoutes(mux RouteRegistrar) {
	mux.HandleFunc("POST /internal/emqx/auth", h.authenticate)
}

func (h *EMQXAuthHandler) authenticate(w http.ResponseWriter, r *http.Request) {
	var req emqxAuthRequest
	if err := decodeJSON(r, &req); err != nil {
		h.writeDeny(w)
		return
	}
	if h.authenticateServiceClient(req) {
		writeJSON(w, http.StatusOK, emqxAuthResponse{Result: "allow", IsSuperuser: true})
		return
	}

	in, ok := mqttAuthInputFromEMQX(req)
	if !ok {
		h.writeDeny(w)
		return
	}

	result, err := h.service.Authenticate(r.Context(), in)
	if err != nil {
		if isMQTTAuthDeny(err) {
			h.writeDeny(w)
			return
		}
		h.log.Error("emqx auth request failed", "err", err)
		writeHTTPError(w, httperror.ErrInternal)
		return
	}

	writeJSON(w, http.StatusOK, emqxAuthResponse{Result: "allow", IsSuperuser: false, ClientAttrs: map[string]string{
		"tenant_id":   result.TenantID,
		"product_id":  result.ProductID,
		"device_id":   result.DeviceID,
		"tenant_slug": result.TenantSlug,
		"product_key": result.ProductKey,
		"device_slug": result.DeviceSlug,
	},
	})
}

func (h *EMQXAuthHandler) writeDeny(w http.ResponseWriter) {
	writeJSON(w, http.StatusOK, emqxAuthResponse{Result: "deny"})
}

func (h *EMQXAuthHandler) authenticateServiceClient(req emqxAuthRequest) bool {
	username := strings.TrimSpace(req.Username)
	password := strings.TrimSpace(req.Password)
	if username != h.serviceUsername {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(password), []byte(h.servicePassword)) == 1
}

func mqttAuthInputFromEMQX(req emqxAuthRequest) (service.MQTTAuthInput, bool) {
	username := strings.TrimSpace(req.Username)
	clientID := strings.TrimSpace(req.ClientID)
	password := strings.TrimSpace(req.Password)
	if username == "" || clientID == "" || password == "" {
		return service.MQTTAuthInput{}, false
	}

	var tenantSlug string
	var productKey string
	var deviceSlug string

	usernameParts := splitIdentity(username)
	switch len(usernameParts) {
	case 1:
		tenantSlug = usernameParts[0]
		clientIDParts := splitIdentity(clientID)
		switch len(clientIDParts) {
		case 1:
			deviceSlug = clientIDParts[0]
		case 2:
			productKey = clientIDParts[0]
			deviceSlug = clientIDParts[1]
		default:
			return service.MQTTAuthInput{}, false
		}
	case 3:
		tenantSlug = usernameParts[0]
		productKey = usernameParts[1]
		deviceSlug = usernameParts[2]
	default:
		return service.MQTTAuthInput{}, false
	}

	return service.MQTTAuthInput{
		TenantSlug: tenantSlug,
		ProductKey: productKey,
		DeviceSlug: deviceSlug,
		Password:   password,
	}, true
}

func splitIdentity(value string) []string {
	raw := strings.Split(value, ":")
	parts := make([]string, 0, len(raw))
	for _, part := range raw {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil
		}
		parts = append(parts, part)
	}
	return parts
}

func isMQTTAuthDeny(err error) bool {
	return errors.Is(err, service.ErrInvalidMQTTAuthInput) || errors.Is(err, service.ErrInvalidMQTTAuth)
}
