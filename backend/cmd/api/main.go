package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/handler"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/middleware"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/repository"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/seed"
	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/service"
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/database"
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/logger"
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/security"
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/storage"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Carrega .env (ignora erro em produção — variáveis já devem estar setadas)
	_ = godotenv.Load()

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}

	// ── Observabilidade: Logger Estruturado JSON (slog) ─
	appLogger := logger.InitLogger(appEnv)
	slog.Info("Iniciando docseq-DMS API...", slog.String("env", appEnv))

	// Validação de configurações de segurança
	if err := security.ValidateJWTConfig(); err != nil {
		slog.Error("Falha de validação das configurações de segurança", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// ── Banco de dados ───────────────────────────────
	db, err := database.Connect()
	if err != nil {
		slog.Error("Falha ao conectar no banco de dados", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := database.Migrate(db); err != nil {
		slog.Error("Falha ao executar migrações do banco", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Garante a existência de um usuário admin padrão (idempotente)
	if err := seed.AdminUser(db); err != nil {
		slog.Error("Falha ao executar seed de admin", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// ── Wiring (DI manual) ──────────────────────────
	municipalityRepo := repository.NewMunicipalityRepository(db)
	municipalitySvc := service.NewMunicipalityService(municipalityRepo)
	municipalityHnd := handler.NewMunicipalityHandler(municipalitySvc)

	permissionRepo := repository.NewUserPermissionRepository(db)
	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(userRepo, municipalityRepo, permissionRepo)
	userHnd := handler.NewUserHandler(userSvc)

	// ── Storage (Cloudflare R2 / S3) ────────────────
	r2Cfg := storage.LoadConfigFromEnv()
	var storageSvc storage.StorageService
	if r2Cfg.IsConfigured() {
		var err error
		storageSvc, err = storage.NewR2StorageService(r2Cfg)
		if err != nil {
			slog.Warn("Falha ao inicializar Cloudflare R2 storage service", slog.String("error", err.Error()))
			storageSvc = storage.NewMockStorageService()
		} else {
			slog.Info("Cloudflare R2 storage conectado com sucesso", slog.String("bucket", r2Cfg.BucketName))
		}
	} else {
		slog.Warn("Cloudflare R2 não configurado. Utilizando MockStorageService para ambiente local")
		storageSvc = storage.NewMockStorageService()
	}

	docRepo := repository.NewDocumentRepository(db)
	docSvc := service.NewDocumentService(docRepo, userRepo, municipalityRepo, storageSvc)
	docHnd := handler.NewDocumentHandler(docSvc, permissionRepo)

	authSvc := service.NewAuthService(userRepo)
	authHnd := handler.NewAuthHandler(authSvc)

	healthHnd := handler.NewHealthHandler(db)

	// ── Router ──────────────────────────────────────
	if appEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.StructuredLoggerMiddleware(appLogger))
	r.Use(middleware.RecoveryWithSlog(appLogger))
	r.Use(middleware.SecurityHeadersMiddleware())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.MaxBodySizeMiddleware(1 << 20)) // Limite de 1MB por payload

	// Health-check ativo com ping de conectividade no banco
	r.GET("/health", healthHnd.HealthCheck)

	// Rate limiter para endpoints sensíveis de autenticação (5 req/min por IP)
	authRateLimiter := middleware.RateLimiterMiddleware(5, 1*time.Minute)

	// API v1
	v1 := r.Group("/api/v1")

	// Rotas públicas (Login com rate limiter)
	authHnd.RegisterRoutes(v1, authRateLimiter)

	// Rotas protegidas por Autenticação JWT
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		municipalityHnd.RegisterRoutes(protected)
		userHnd.RegisterRoutes(protected, authRateLimiter)
		docHnd.RegisterRoutes(protected)
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
	}

	// Executa o servidor em goroutine separada
	go func() {
		slog.Info("🚀 docseq-DMS rodando com sucesso", slog.String("port", port), slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Erro inesperado no servidor HTTP", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// ── Graceful Shutdown ───────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	sig := <-quit
	slog.Info("🛑 Sinal de encerramento recebido. Encerrando servidor graciosamente...", slog.String("signal", sig.String()))

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Servidor forçado a encerrar antes de finalizar requisições em voo", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if sqlDB, err := db.DB(); err == nil {
		if closeErr := sqlDB.Close(); closeErr != nil {
			slog.Error("Erro ao encerrar conexões com o pool do banco", slog.String("error", closeErr.Error()))
		} else {
			slog.Info("Conexões com o banco de dados encerradas com sucesso")
		}
	}

	slog.Info("✅ Servidor docseq-DMS finalizado com sucesso.")
}
