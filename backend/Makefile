.PHONY: test test-unit test-integration test-coverage

# ── Rodar todos os testes ────────────────────────────────────────────────────
test:
	go test ./... -v

# ── Apenas testes unitários (service + handler — sem Docker) ─────────────────
test-unit:
	go test ./internal/service/... ./internal/handler/... -v -count=1

# ── Testes de integração do repositório (exige Docker) ───────────────────────
test-integration:
	go test ./internal/repository/... -v -count=1 -timeout 120s

# ── Cobertura HTML ───────────────────────────────────────────────────────────
test-coverage:
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "Relatório gerado: coverage.html"

# ── Dependências de teste ────────────────────────────────────────────────────
test-deps:
	go get github.com/stretchr/testify@latest
	go get github.com/testcontainers/testcontainers-go@latest
