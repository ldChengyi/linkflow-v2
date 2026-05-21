package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/ldchengyi/linkflow-v2/service/backend/internal/httperror"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/middleware"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/response"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/service"
)

const (
	realtimeAccessCookieName = "lf_access"
	realtimeAccessCookiePath = "/api/v1/ws"
)

type AuthHandler struct {
	service           *service.AuthService
	accessTokenMaxAge int
	log               *slog.Logger
}

type authRegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewAuthHandler(service *service.AuthService, accessTokenTTL time.Duration, log *slog.Logger) (*AuthHandler, error) {
	if service == nil {
		return nil, errors.New("auth service is nil")
	}
	if accessTokenTTL <= 0 {
		return nil, errors.New("access token ttl must be positive")
	}
	if log == nil {
		log = slog.Default()
	}
	return &AuthHandler{
		service:           service,
		accessTokenMaxAge: int(accessTokenTTL.Seconds()),
		log:               log,
	}, nil
}

func (h *AuthHandler) RegisterRoutes(mux RouteRegistrar, authenticate func(http.Handler) http.Handler) {
	mux.HandleFunc("POST /api/v1/auth/register", h.register)
	mux.HandleFunc("POST /api/v1/auth/login", h.login)
	if authenticate != nil {
		mux.Handle("POST /api/v1/auth/logout", authenticate(http.HandlerFunc(h.logout)))
	}
}

func (h *AuthHandler) register(w http.ResponseWriter, r *http.Request) {
	var req authRegisterRequest
	if err := decodeJSON(r, &req); err != nil {
		writeHTTPError(w, err)
		return
	}

	result, err := h.service.Register(r.Context(), service.AuthRegisterInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		h.writeAuthError(w, err)
		return
	}

	h.setRealtimeAccessCookie(w, result.AccessToken)
	writeJSON(w, http.StatusCreated, response.SuccessData("registered", http.StatusCreated, result))
}

func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	var req authLoginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeHTTPError(w, err)
		return
	}

	result, err := h.service.Login(r.Context(), service.AuthLoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		h.writeAuthError(w, err)
		return
	}

	h.setRealtimeAccessCookie(w, result.AccessToken)
	writeJSON(w, http.StatusOK, response.SuccessData("ok", http.StatusOK, result))
}

func (h *AuthHandler) logout(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		writeHTTPError(w, httperror.ErrUnauthorized)
		return
	}

	if err := h.service.Logout(r.Context(), service.AuthLogoutInput{
		TokenID: principal.TokenID,
		UserID:  principal.UserID,
	}); err != nil {
		h.writeAuthError(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     realtimeAccessCookieName,
		Value:    "",
		Path:     realtimeAccessCookiePath,
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	writeJSON(w, http.StatusOK, response.SuccessData("logged out", http.StatusOK, map[string]bool{"revoked": true}))
}

func (h *AuthHandler) setRealtimeAccessCookie(w http.ResponseWriter, accessToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     realtimeAccessCookieName,
		Value:    accessToken,
		Path:     realtimeAccessCookiePath,
		MaxAge:   h.accessTokenMaxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *AuthHandler) writeAuthError(w http.ResponseWriter, err error) {
	httpErr := mapAuthError(err)
	if httpErr.HTTPStatus >= http.StatusInternalServerError {
		h.log.Error("auth request failed", "err", err)
	}
	writeHTTPError(w, httpErr)
}

func mapAuthError(err error) *httperror.Error {
	switch {
	case errors.Is(err, service.ErrInvalidAuthInput):
		return httperror.ErrInvalidRequest
	case errors.Is(err, service.ErrInvalidCredentials):
		return httperror.ErrInvalidCredentials
	case errors.Is(err, service.ErrUserAlreadyExists):
		return httperror.ErrUserAlreadyExists
	case errors.Is(err, service.ErrUserDisabled):
		return httperror.ErrUserDisabled
	default:
		return httperror.From(err)
	}
}
