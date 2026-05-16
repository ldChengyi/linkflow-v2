package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	publicpostgres "github.com/ldchengyi/linkflow-v2/pkg/public/postgres"
	publicredis "github.com/ldchengyi/linkflow-v2/pkg/public/redis"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/auth/password"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/auth/token"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/config"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/credential"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/handler"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/middleware"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/server"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/service"
	"github.com/ldchengyi/linkflow-v2/service/backend/internal/store"
)

const shutdownTimeout = 10 * time.Second

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, log); err != nil {
		log.Error("backend stopped", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, log *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	tokenManager, err := token.NewJWTManager(token.Config{
		Secret: cfg.AuthAccessTokenSecret,
		TTL:    cfg.AuthAccessTokenTTL,
		Issuer: cfg.AuthTokenIssuer,
	})
	if err != nil {
		return fmt.Errorf("create jwt manager: %w", err)
	}

	postgresPool, err := publicpostgres.Open(ctx, publicpostgres.Config{
		DSN:             cfg.PostgresDSN,
		MaxConns:        int32(cfg.PostgresMaxOpenConns),
		MinConns:        int32(cfg.PostgresMaxIdleConns),
		MaxConnLifetime: cfg.PostgresConnMaxLifetime,
		MaxConnIdleTime: cfg.PostgresConnMaxIdleTime,
	})
	if err != nil {
		return fmt.Errorf("open postgres: %w", err)
	}
	defer postgresPool.Close()

	redisClient, err := publicredis.Open(ctx, publicredis.Config{
		Addr:         cfg.RedisAddr,
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		DialTimeout:  cfg.RedisDialTimeout,
		ReadTimeout:  cfg.RedisReadTimeout,
		WriteTimeout: cfg.RedisWriteTimeout,
	})
	if err != nil {
		return fmt.Errorf("open redis: %w", err)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Warn("close redis", "err", err)
		}
	}()

	userStore, err := store.NewPostgresAuthStore(postgresPool)
	if err != nil {
		return fmt.Errorf("create auth user store: %w", err)
	}
	sessionStore, err := store.NewRedisAuthSessionStore(redisClient)
	if err != nil {
		return fmt.Errorf("create auth session store: %w", err)
	}
	passwordHasher, err := password.NewBcryptHasher(cfg.AuthBCryptCost)
	if err != nil {
		return fmt.Errorf("create password hasher: %w", err)
	}
	deviceSecretManager, err := credential.NewSecretManager(cfg.AuthBCryptCost)
	if err != nil {
		return fmt.Errorf("create device secret manager: %w", err)
	}
	authService, err := service.NewAuthService(userStore, sessionStore, passwordHasher, tokenManager)
	if err != nil {
		return fmt.Errorf("create auth service: %w", err)
	}
	authHandler, err := handler.NewAuthHandler(authService, log)
	if err != nil {
		return fmt.Errorf("create auth handler: %w", err)
	}
	tenantStore, err := store.NewPostgresTenantStore(postgresPool)
	if err != nil {
		return fmt.Errorf("create tenant store: %w", err)
	}
	tenantService, err := service.NewTenantService(tenantStore)
	if err != nil {
		return fmt.Errorf("create tenant service: %w", err)
	}
	tenantHandler, err := handler.NewTenantHandler(tenantService, log)
	if err != nil {
		return fmt.Errorf("create tenant handler: %w", err)
	}
	productStore, err := store.NewPostgresProductStore(postgresPool)
	if err != nil {
		return fmt.Errorf("create product store: %w", err)
	}
	productService, err := service.NewProductService(productStore)
	if err != nil {
		return fmt.Errorf("create product service: %w", err)
	}
	productHandler, err := handler.NewProductHandler(productService, log)
	if err != nil {
		return fmt.Errorf("create product handler: %w", err)
	}
	thingsModelStore, err := store.NewPostgresThingsModelStore(postgresPool)
	if err != nil {
		return fmt.Errorf("create thingsmodel store: %w", err)
	}
	thingsModelService, err := service.NewThingsModelService(thingsModelStore)
	if err != nil {
		return fmt.Errorf("create thingsmodel service: %w", err)
	}
	thingsModelHandler, err := handler.NewThingsModelHandler(thingsModelService, log)
	if err != nil {
		return fmt.Errorf("create thingsmodel handler: %w", err)
	}
	deviceStore, err := store.NewPostgresDeviceStore(postgresPool)
	if err != nil {
		return fmt.Errorf("create device store: %w", err)
	}
	deviceService, err := service.NewDeviceService(deviceStore, deviceSecretManager)
	if err != nil {
		return fmt.Errorf("create device service: %w", err)
	}
	deviceHandler, err := handler.NewDeviceHandler(deviceService, log)
	if err != nil {
		return fmt.Errorf("create device handler: %w", err)
	}
	auditStore, err := store.NewPostgresAuditStore(postgresPool)
	if err != nil {
		return fmt.Errorf("create audit store: %w", err)
	}
	auditService, err := service.NewAuditService(auditStore)
	if err != nil {
		return fmt.Errorf("create audit service: %w", err)
	}
	auditHandler, err := handler.NewAuditHandler(auditService, log)
	if err != nil {
		return fmt.Errorf("create audit handler: %w", err)
	}

	httpServer, err := server.New(cfg, log, server.Options{
		Auth:         authHandler,
		Tenant:       tenantHandler,
		Product:      productHandler,
		ThingsModel:  thingsModelHandler,
		Device:       deviceHandler,
		AuditLog:     auditHandler,
		Authenticate: middleware.Authenticate(tokenManager, sessionStore),
		Audit:        middleware.Audit(auditService, log),
	})
	if err != nil {
		return fmt.Errorf("create http server: %w", err)
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("backend started", "addr", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown http server: %w", err)
		}
		return nil
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("run http server: %w", err)
		}
		return nil
	}
}
