package main

import (
	"context"
	"errors"
	"log"
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
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/security"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Carrega .env (ignora erro em produção — variáveis já devem estar setadas)
	_ = godotenv.Load()

	// Validação de configurações de segurança
	if err := security.ValidateJWTConfig(); err != nil {
		log.Fatalf("segurança: %v", err)
	}

	// ── Banco de dados ───────────────────────────────
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	// Garante a existência de um usuário admin padrão (idempotente)
	if err := seed.AdminUser(db); err != nil {
		log.Fatalf("seed: %v", err)
	}

	// ── Wiring (DI manual) ──────────────────────────
	municipalityRepo := repository.NewMunicipalityRepository(db)
	municipalitySvc := service.NewMunicipalityService(municipalityRepo)
	municipalityHnd := handler.NewMunicipalityHandler(municipalitySvc)

	permissionRepo := repository.NewUserPermissionRepository(db)
	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(userRepo, municipalityRepo, permissionRepo)
	userHnd := handler.NewUserHandler(userSvc)

	docRepo := repository.NewDocumentRepository(db)
	docSvc := service.NewDocumentService(docRepo, userRepo, municipalityRepo)
	docHnd := handler.NewDocumentHandler(docSvc, permissionRepo)

	authSvc := service.NewAuthService(userRepo)
	authHnd := handler.NewAuthHandler(authSvc)

	// ── Router ──────────────────────────────────────
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.Use(middleware.SecurityHeadersMiddleware())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.MaxBodySizeMiddleware(1 << 20)) // Limite de 1MB por payload

	// Health-check ativo com validação de conectividade do banco
	r.GET("/health", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "database": "unreachable"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := sqlDB.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "database": "down"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok", "database": "up"})
	})

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
		log.Printf("🚀  docseq-DMS rodando em :%s", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server listen: %v", err)
		}
	}()

	// ── Graceful Shutdown ───────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Sinal de encerramento recebido. Encerrando servidor graciosamente...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Servidor forçado a encerrar: %v", err)
	}

	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}

	log.Println("✅ Servidor docseq-DMS finalizado com sucesso.")
}
