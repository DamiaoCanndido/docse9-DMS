package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/seu-usuario/doc-manager/internal/handler"
	"github.com/seu-usuario/doc-manager/internal/repository"
	"github.com/seu-usuario/doc-manager/internal/service"
	"github.com/seu-usuario/doc-manager/pkg/database"
)

func main() {
	// Carrega .env (ignora erro em produção — variáveis já devem estar setadas)
	_ = godotenv.Load()

	// ── Banco de dados ───────────────────────────────
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	// ── Wiring (DI manual) ──────────────────────────
	municipalityRepo := repository.NewMunicipalityRepository(db)
	municipalitySvc  := service.NewMunicipalityService(municipalityRepo)
	municipalityHnd  := handler.NewMunicipalityHandler(municipalitySvc)

	// ── Router ──────────────────────────────────────
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Health-check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	v1 := r.Group("/api/v1")
	municipalityHnd.RegisterRoutes(v1)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀  doc-manager rodando em :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server: %v", err)
	}
}
