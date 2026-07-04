package seed

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/security"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	defaultAdminUsername    = "admin"
	defaultAdminEmail       = "admin@docse9.local"
	defaultAdminPassword    = "admin123"
	defaultMunicipalityName = "Matriz"
	defaultMunicipalityUF   = "PB"
)

// AdminUser garante que exista um usuário administrador padrão no banco.
// É idempotente: pode ser chamada em todo boot da aplicação sem duplicar
// o usuário nem sobrescrever um admin já customizado pelo operador.
//
// Variáveis de ambiente suportadas (todas opcionais, com fallback seguro):
//
//	DEFAULT_ADMIN_USERNAME
//	DEFAULT_ADMIN_EMAIL
//	DEFAULT_ADMIN_PASSWORD
//	DEFAULT_MUNICIPALITY_NAME (usado só se ainda não existir nenhum município)
//	DEFAULT_MUNICIPALITY_UF
func AdminUser(db *gorm.DB) error {
	username := envOrDefault("DEFAULT_ADMIN_USERNAME", defaultAdminUsername)
	email := envOrDefault("DEFAULT_ADMIN_EMAIL", defaultAdminEmail)
	password := envOrDefault("DEFAULT_ADMIN_PASSWORD", defaultAdminPassword)

	if os.Getenv("APP_ENV") == "production" && os.Getenv("DEFAULT_ADMIN_PASSWORD") == "" {
		return errors.New("DEFAULT_ADMIN_PASSWORD não definida em produção — defina uma senha forte antes de subir a aplicação")
	}

	exists, err := adminAlreadyExists(db, username, email)
	if err != nil {
		return fmt.Errorf("seed: checar usuário admin: %w", err)
	}
	if exists {
		log.Println("🌱  Usuário admin padrão já existe — seed ignorado")
		return nil
	}

	municipalityID, err := ensureDefaultMunicipality(db)
	if err != nil {
		return fmt.Errorf("seed: garantir município padrão: %w", err)
	}

	hash, err := security.HashPassword(password)
	if err != nil {
		return fmt.Errorf("seed: gerar hash da senha: %w", err)
	}

	admin := domain.User{
		Username:       username,
		Email:          email,
		Password:       hash,
		Role:           domain.RoleAdmin,
		MunicipalityID: municipalityID,
	}

	if err := db.Create(&admin).Error; err != nil {
		return fmt.Errorf("seed: criar usuário admin: %w", err)
	}

	log.Printf("🌱  Usuário admin padrão criado (username=%q, email=%q)\n", username, email)
	if os.Getenv("DEFAULT_ADMIN_PASSWORD") == "" {
		log.Println("⚠️   DEFAULT_ADMIN_PASSWORD não foi definida — senha padrão de fábrica em uso (apenas ambientes não-produtivos). Troque-a assim que possível.")
	}

	return nil
}

// adminAlreadyExists verifica, inclusive entre registros soft-deletados, se já
// existe um usuário com o username ou e-mail informados — evitando tanto a
// duplicação quanto uma tentativa de recriação que colidiria com o índice
// único do banco.
func adminAlreadyExists(db *gorm.DB, username, email string) (bool, error) {
	var count int64
	err := db.Unscoped().
		Model(&domain.User{}).
		Where("LOWER(username) = LOWER(?) OR LOWER(email) = LOWER(?)", username, email).
		Count(&count).Error
	return count > 0, err
}

// ensureDefaultMunicipality retorna o ID de um município ao qual vincular o
// admin padrão, criando um município "seed" caso nenhum exista ainda.
func ensureDefaultMunicipality(db *gorm.DB) (uuid.UUID, error) {
	var m domain.Municipality
	err := db.First(&m).Error
	if err == nil {
		return m.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return uuid.Nil, err
	}

	m = domain.Municipality{
		Name: envOrDefault("DEFAULT_MUNICIPALITY_NAME", defaultMunicipalityName),
		UF:   envOrDefault("DEFAULT_MUNICIPALITY_UF", defaultMunicipalityUF),
	}
	if err := db.Create(&m).Error; err != nil {
		return uuid.Nil, err
	}

	log.Printf("🌱  Município padrão criado (%s/%s) para vincular o admin\n", m.Name, m.UF)
	return m.ID, nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
