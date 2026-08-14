package main

import (
	"log"
	"os"

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

	// Health-check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	v1 := r.Group("/api/v1")

	// Rotas públicas (Login)
	authHnd.RegisterRoutes(v1)

	// Rotas protegidas por Autenticação JWT
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		municipalityHnd.RegisterRoutes(protected)
		userHnd.RegisterRoutes(protected)
		docHnd.RegisterRoutes(protected)
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀  docseq-DMS rodando em :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server: %v", err)
	}
}
