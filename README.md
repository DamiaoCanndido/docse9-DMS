# 🏛️ Docseq DMS — Sistema de Gestão de Documentos Oficiais

<div align="center">

[![Go Version](https://img.shields.io/badge/Go-1.25-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org)
[![Next.js](https://img.shields.io/badge/Next.js-16-000000?style=for-the-badge&logo=next.js&logoColor=white)](https://nextjs.org)
[![React](https://img.shields.io/badge/React-19-61DAFB?style=for-the-badge&logo=react&logoColor=black)](https://react.dev)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=for-the-badge&logo=postgresql&logoColor=white)](https://www.postgresql.org)
[![Tailwind CSS v4](https://img.shields.io/badge/Tailwind_CSS-v4-38B2AC?style=for-the-badge&logo=tailwind-css&logoColor=white)](https://tailwindcss.com)
[![Google Cloud Run](https://img.shields.io/badge/GCP-Cloud_Run-4285F4?style=for-the-badge&logo=google-cloud&logoColor=white)](https://cloud.google.com/run)
[![Netlify](https://img.shields.io/badge/Netlify-Deployed-00C7B7?style=for-the-badge&logo=netlify&logoColor=white)](https://www.netlify.com)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge)](LICENSE)

<p align="center">
  <strong>Plataforma moderna, multi-tenant e auditável para criação, numeração sequencial, controle de acesso e arquivamento digital de documentos públicos municipais.</strong>
</p>

</div>

---

## 📌 Sumário

- [Visão Geral](#-visão-geral)
- [Principais Funcionalidades](#-principais-funcionalidades)
- [Arquitetura do Sistema](#-arquitetura-do-sistema)
- [Modelo de Permissões e Segurança (RBAC)](#-modelo-de-permissões-e-segurança-rbac)
- [Tecnologias Utilizadas](#-tecnologias-utilizadas)
- [Estrutura do Projeto](#-estrutura-do-projeto)
- [Guia de Início Rápido (Desenvolvimento Local)](#-guia-de-início-rápido-desenvolvimento-local)
- [Testes Automatizados e Qualidade](#-testes-automatizados-e-qualidade)
- [Deploy e CI/CD](#-deploy-e-cicd)
- [Ciclo de Desenvolvimento com IA Multi-Agentes](#-ciclo-de-desenvolvimento-com-ia-multi-agentes)
- [Licença](#-licença)

---

## 📖 Visão Geral

O **Docseq DMS** (*Document Management System*) é uma solução web e API desenvolvida para atender às exigências de transparência, integridade e governança de órgãos públicos municipais (prefeituras, câmaras de vereadores, secretarias e autarquias).

Projetado com arquitetura **Multi-Tenant** e padrão de isolamento por município, o sistema permite que múltiplos entes federativos operem na mesma infraestrutura garantindo segregação total de dados, controle rigoroso de acesso e trilhas de auditoria para cada documento emitido.

---

## ✨ Principais Funcionalidades

- **🏢 Multi-Inquilinato (Multi-Tenancy):** Separação lógica de dados por Município (`MunicipalityID`).
- **📄 Gestão de Tipos de Documentos Oficiais:**
  - **Ofícios (`NOTICE`):** Comunicações oficiais internas e externas. Numeração reiniciada anualmente.
  - **Decretos (`DECREE`):** Atos normativos do poder executivo. Numeração reiniciada anualmente.
  - **Portarias (`ORDINANCE`):** Instruções e determinações de secretarias/órgãos. Numeração reiniciada anualmente.
  - **Leis (`LAW`):** Legislação municipal. Numeração sequencial perpétua (não reinicia ao virar o ano).
  - **Contratos (`CONTRACT`):** Gestão de contratos (Interesse Público, Licitação, Prestação de Serviços) com controle de vigência, datas, valor e numeração anual.
- **🔢 Controle de Sequência e Numeração Automática:** Garantia transacional de geração sequencial do número de ordem (`order`) por tipo e município.
- **🗑️ Ciclo de Vida e Lixeira (Soft Delete):** Suporte a exclusão lógica, visualização em lixeira isolada, restauração e exclusão definitiva (*Hard Delete*) por perfis autorizados.
- **🔒 Segurança e Conformidade com a LGPD:** Criptografia de senhas com `bcrypt`, autenticação stateless via `JWT Bearer`, cabeçalhos HTTP de segurança rigorosos e política de primeiro acesso com troca obrigatória de senha (`MustChangePassword`).
- **👥 Controle de Acesso Granular (RBAC):** Níveis de permissão por função e matriz de autorização por tipo de documento.

---

## 🏗️ Arquitetura do Sistema

### Arquitetura em Camadas do Backend (Clean Architecture / DDD)

O backend segue a separação estrita de responsabilidades com dependências apontando para o domínio central:

```mermaid
flowchart TD
    subgraph Client["Clientes / Frontend"]
        SPA["Next.js 16 Web App"]
        API_REQ["Requisições HTTP / REST"]
    end

    subgraph HTTP_Layer["Camada de Entrega (HTTP & Middleware)"]
        Router["Gin Router (/api/v1)"]
        AuthMid["Auth & Role Middlewares"]
        SecMid["Security Headers & CORS"]
        Handlers["HTTP Handlers\n(Auth, Municipality, User, Document)"]
    end

    subgraph Service_Layer["Camada de Aplicação (Lógica de Negócio)"]
        Services["Services\n(Auth, Municipality, User, Document)"]
        Validation["Validações de Domínio & Hashing"]
    end

    subgraph Repository_Layer["Camada de Persistência"]
        Repos["GORM Repositories\n(PostgreSQL Adapter)"]
    end

    subgraph Domain_Layer["Camada de Domínio"]
        Entities["Entidades & Interfaces\n(Municipality, User, Document, Permission)"]
    end

    subgraph DB["Infraestrutura de Dados"]
        Postgres[(PostgreSQL 16\nDatabase)]
    end

    SPA --> API_REQ
    API_REQ --> SecMid --> AuthMid --> Router --> Handlers
    Handlers --> Services
    Services --> Entities
    Services --> Repos
    Repos --> Postgres
```

### Arquitetura de Nuvem e Deploy

```mermaid
flowchart LR
    Dev[Desenvolvedor / Git] -->|Push na branch master| GHA[GitHub Actions CI/CD]
    
    subgraph CI_Pipeline["Pipeline CI / CD"]
        GHA -->|Executa Testes Unitários & Lint| TestJob[Test & Quality Gate]
        TestJob -->|Build Docker Multi-Stage| DockerBuild[Docker Buildx]
        DockerBuild -->|Push Imagem Taggeada| GAR[Google Artifact Registry]
        DockerBuild -->|Deploy Automático| CloudRun[GCP Cloud Run (Go API)]
    end

    subgraph Cloud_Infra["Infraestrutura de Produção"]
        CloudRun -->|Injeta Segredos em Runtime| SecretMgr[GCP Secret Manager]
        CloudRun -->|Conexão Pooling via TLS| NeonDB[(Neon / Supabase PostgreSQL)]
        NetlifyApp[Netlify Edge CDN] -->|Consome API REST| CloudRun
    end
```

### Modelo de Entidades do Banco de Dados

```mermaid
erDiagram
    MUNICIPALITY ||--o{ USER : "possui"
    MUNICIPALITY ||--o{ DOCUMENT : "pertence a"
    USER ||--o{ DOCUMENT : "cria"
    USER ||--o{ USER_PERMISSION : "possui permissões"

    MUNICIPALITY {
        uuid id PK
        varchar name UK
        char uf
        text imageUrl
        timestamp createdAt
        timestamp updatedAt
        timestamp deletedAt
    }

    USER {
        uuid id PK
        varchar username UK
        varchar email UK
        varchar password
        varchar role "ADMIN | MOD | COMMON"
        boolean mustChangePassword
        uuid municipalityId FK
        timestamp lastLogin
        timestamp createdAt
        timestamp updatedAt
        timestamp deletedAt
    }

    USER_PERMISSION {
        uuid id PK
        uuid userId FK
        varchar documentType "NOTICE | DECREE | ORDINANCE | LAW | CONTRACT"
        varchar level "NONE | READ | WRITE | DELETE"
        timestamp createdAt
        timestamp updatedAt
    }

    DOCUMENT {
        uuid id PK
        varchar type "NOTICE | DECREE | ORDINANCE | LAW | CONTRACT"
        integer order
        text description
        varchar fileKey
        uuid creatorId FK
        uuid municipalityId FK
        integer duration
        varchar contractType "publicinterest | bidding | service"
        numeric value
        timestamp startIn
        timestamp createdAt
        timestamp updatedAt
        timestamp deletedAt
    }
```

---

## 🔐 Modelo de Permissões e Segurança (RBAC)

O Docseq DMS opera com 3 papéis principais com escopos rigorosamente delimitados:

| Papel (`Role`) | Gestão de Municípios | Gestão de Usuários | Gestão de Documentos |
| :--- | :---: | :---: | :---: |
| **`ADMIN`** (Super Administrador) | Total (Criar, Listar, Editar, Excluir) | Criar/Gerenciar Moderadores (`MOD`) e Usuários de qualquer município | Acesso restrito apenas a tarefas de sistema (não gerencia documentos de negócio) |
| **`MOD`** (Moderador Municipal) | Somente Leitura do seu município | Criar e Gerenciar usuários (`COMMON`) e moderadores do seu próprio município | Acesso Total a todos os tipos de documentos do seu município |
| **`COMMON`** (Operador Municipal) | Sem acesso | Sem acesso (apenas gerencia o próprio perfil/senha) | Conforme Matriz Granular de Permissões configurada pelo `MOD` |

### Matriz de Permissões Granulares para Usuários `COMMON`

Cada usuário comum pode receber autorizações específicas por tipo de documento:

- **`NONE` (0):** Sem permissão de acesso ao tipo de documento.
- **`READ` (1):** Visualizar e pesquisar documentos do tipo.
- **`WRITE` (2):** Visualizar, pesquisar, criar e editar documentos do tipo.
- **`DELETE` (3):** Visualizar, pesquisar, criar, editar, enviar para a lixeira, restaurar e excluir definitivamente.

---

## 🛠️ Tecnologias Utilizadas

### Backend
- **Linguagem:** [Go 1.25](https://golang.org)
- **Framework Web:** [Gin Gonic (v1.12)](https://github.com/gin-gonic/gin)
- **ORM & Banco de Dados:** [GORM (v1.25)](https://gorm.io) com driver PostgreSQL [pgx (v5)](https://github.com/jackc/pgx)
- **Autenticação:** [JWT (golang-jwt/jwt/v5)](https://github.com/golang-jwt/jwt) e Criptografia com [bcrypt (x/crypto)](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
- **Testes & Mocks:** [Testify](https://github.com/stretchr/testify) e [Testcontainers-Go](https://golang.testcontainers.org/)
- **Documentação de API:** OpenAPI 3.0 (`openapi.yaml`)

### Frontend
- **Framework:** [Next.js 16](https://nextjs.org) (App Router com Server Components & Client Components)
- **Biblioteca UI:** [React 19](https://react.dev) + [TypeScript](https://www.typescriptlang.org)
- **Estilização:** [Tailwind CSS v4](https://tailwindcss.com) + [Lucide Icons](https://lucide.dev) + [Framer Motion](https://www.framer.com/motion/) + [Shadcn UI](https://ui.shadcn.com)
- **Formulários e Validação:** [React Hook Form](https://react-hook-form.com) + [Zod](https://zod.dev)
- **Requisições HTTP:** [Axios](https://axios-http.com) com interceptors de autenticação
- **Notificações:** [Sonner](https://sonner.emilkowal.ski)

### Infraestrutura & DevOps
- **Containerização:** Docker e Docker Compose
- **Computação em Nuvem:** Google Cloud Run (Serverless Container)
- **Gerenciamento de Segredos:** GCP Secret Manager
- **Hospedagem de Frontend:** Netlify
- **Integração e Entrega Contínua (CI/CD):** GitHub Actions

---

## 📁 Estrutura do Projeto

```
docSe9-DMS/
├── .agents/                    # Definições de agentes de IA, skills e workflows
│   ├── skills/                 # Skills especializadas (write-specs, generate-code, audit-code, etc.)
│   └── workflows/              # Fluxos de orquestração (/startcycle)
├── .github/
│   └── workflows/
│       ├── ci.yml              # Pipeline de CI (Lint, Go Vet, Testes Unitários)
│       └── cd-gcp.yml          # Pipeline de CD (Build Docker, Artifact Registry, Deploy Cloud Run)
├── backend/                    # Código-fonte da API em Go
│   ├── cmd/api/main.go         # Ponto de entrada da aplicação e Injeção de Dependências
│   ├── internal/
│   │   ├── domain/             # Entidades, DTOs e Interfaces de Repositório/Serviço
│   │   ├── handler/            # Controllers HTTP e validação de requisições
│   │   ├── middleware/         # Middlewares (Auth JWT, RBAC, CORS, Security Headers)
│   │   ├── repository/         # Implementações de acesso ao banco com GORM
│   │   ├── seed/               # Seed automático do usuário administrador inicial
│   │   └── service/            # Regras de negócio e casos de uso
│   ├── pkg/
│   │   ├── database/           # Pool de conexões PostgreSQL e AutoMigrations
│   │   ├── response/           # Padronização de respostas JSON (Envelope Pattern)
│   │   └── security/           # Utilitários de Token JWT e Hashing
│   ├── docker-compose.yml      # Banco de dados PostgreSQL para desenvolvimento local
│   ├── Dockerfile              # Imagem Docker multi-stage otimizada para produção
│   ├── Makefile                # Atalhos para rodar testes e cobertura
│   └── openapi.yaml            # Especificação completa da API REST (OpenAPI 3.0)
├── docs/                       # Documentações técnicas e guias operacionais
│   └── deploy-gcp-cloud-run.md # Passo a passo para infraestrutura no GCP e Neon/Supabase
├── frontend/                   # Aplicação Web em Next.js 16
│   ├── src/
│   │   ├── app/                # Rotas da aplicação (Auth, Dashboard, Municípios, Usuários, etc.)
│   │   ├── components/         # Componentes reutilizáveis e layouts
│   │   ├── context/            # Contextos React (AuthContext, ThemeProvider)
│   │   ├── lib/                # Configurações de API client (Axios), utilitários e Zod schemas
│   │   └── types/              # Definições de tipos TypeScript
│   ├── netlify.toml            # Configuração de build e deploy no Netlify
│   └── package.json            # Dependências e scripts do frontend
├── netlify.toml                # Configuração raiz para deploy do frontend
└── README.md                   # Este documento
```

---

## 🚀 Guia de Início Rápido (Desenvolvimento Local)

### 📋 Pré-requisitos
- [Git](https://git-scm.com)
- [Go 1.25+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org) e `npm`
- [Docker](https://www.docker.com) e [Docker Compose](https://docs.docker.com/compose/)

---

### 1️⃣ Clonar o Repositório
```bash
git clone https://github.com/DamiaoCanndido/docse9-DMS.git
cd docse9-DMS
```

---

### 2️⃣ Iniciar o Banco de Dados Local (PostgreSQL)
Execute o container do PostgreSQL configurado no docker-compose do backend:

```bash
docker compose -f backend/docker-compose.yml up -d
```
> [!NOTE]
> O banco estará acessível na porta local `5455` com o banco `docseq_dms` e usuário `postgres`.

---

### 3️⃣ Configurar e Rodar o Backend

1. Acesse a pasta do backend:
   ```bash
   cd backend
   ```
2. Crie o arquivo `.env` a partir do modelo:
   ```bash
   cp .env.example .env
   ```
3. Exemplo de configuração para o `.env` local:
   ```env
   APP_PORT=8080
   APP_ENV=development
   DATABASE_URL=postgres://postgres:postgres@localhost:5455/docseq_dms?sslmode=disable
   JWT_SECRET=super_secret_jwt_key_for_development_only_12345!
   CORS_ALLOWED_ORIGINS=http://localhost:3000

   # Configurações do Administrador Padrão (Criado automaticamente se o banco estiver vazio)
   DEFAULT_ADMIN_USERNAME=admin
   DEFAULT_ADMIN_EMAIL=admin@docseq.local
   DEFAULT_ADMIN_PASSWORD=AdminPassword123!
   DEFAULT_MUNICIPALITY_NAME=Município Padrão
   DEFAULT_MUNICIPALITY_UF=PB
   ```
4. Inicie o servidor Go:
   ```bash
   go run ./cmd/api
   ```
   *As migrações do banco e a criação do usuário administrador padrão serão executadas automaticamente ao inicializar.*
   *O backend estará respondendo em `http://localhost:8080` (verifique em `http://localhost:8080/health`).*

---

### 4️⃣ Configurar e Rodar o Frontend

1. Em um novo terminal, acesse a pasta do frontend:
   ```bash
   cd frontend
   ```
2. Crie o arquivo `.env`:
   ```env
   NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
   ```
3. Instale as dependências e inicie o servidor de desenvolvimento:
   ```bash
   npm install
   npm run dev
   ```
4. Acesse no seu navegador: **[http://localhost:3000](http://localhost:3000)**

---

### 🔑 Credenciais Padrão para Login

Ao inicializar o sistema pela primeira vez com o seed, utilize:
- **E-mail:** `admin@docseq.local` (ou usuário: `admin`)
- **Senha:** `AdminPassword123!`
- **Perfil:** `ADMIN`

> [!TIP]
> Por padrão de segurança, o sistema exigirá a alteração de senha no primeiro acesso (`MustChangePassword = true`).

---

## 🧪 Testes Automatizados e Qualidade

O projeto possui uma suíte abrangente de testes automatizados no backend e validações estáticas no frontend.

### Executando Testes do Backend (via Makefile)

Acesse a pasta `backend/`:

```bash
# 1. Executar apenas os testes unitários (Services e Handlers com mocks rápidos - sem Docker)
make test-unit

# 2. Executar testes de integração (Repositories com banco real via Testcontainers - exige Docker)
make test-integration

# 3. Executar todos os testes do backend
make test

# 4. Gerar relatório visual de cobertura de testes (HTML)
make test-coverage
```

### Executando Verificações no Frontend

Acesse a pasta `frontend/`:

```bash
# Executar o linter (ESLint)
npm run lint

# Executar a verificação de tipos e build de produção
npm run build
```

---

## ☁️ Deploy e CI/CD

### Backend (GCP Cloud Run + Neon PostgreSQL)

O repositório está configurado para deploy automático via **GitHub Actions** (`.github/workflows/cd-gcp.yml`):

1. **Trigger:** Push na branch `master` modificando arquivos em `backend/`.
2. **Build:** Constrói uma imagem Docker multi-stage minimalista (Scratch/Distroless).
3. **Artifacts:** Publica a imagem no **Google Artifact Registry (GAR)** com a tag do commit (`${{ github.sha }}`) e `latest`.
4. **Deploy:** Atualiza o serviço no **GCP Cloud Run** com variáveis injetadas via **GCP Secret Manager** (`DOCSEQ_DATABASE_URL`, `DOCSEQ_JWT_SECRET`, etc.).
5. **Health Check:** Valida a resposta do endpoint `/health` antes de finalizar o workflow.

> [!NOTE]
> Consulte o guia detalhado em [`docs/deploy-gcp-cloud-run.md`](docs/deploy-gcp-cloud-run.md) para o passo a passo de configuração do projeto GCP, permissões de IAM e banco de dados Neon/Supabase.

### Procedimento de Rollback Rápido

Caso uma nova versão apresente anomalias em produção:

```bash
# Listar as revisões disponíveis no Cloud Run
gcloud run revisions list --service=docseq-backend-api --region=southamerica-east1

# Redirecionar instantaneamente 100% do tráfego para a revisão anterior estável:
gcloud run services update-traffic docseq-backend-api \
  --region=southamerica-east1 \
  --to-revisions=docseq-backend-api-00001-abc=100
```

### Frontend (Netlify)

O frontend é construído e publicado na CDN global do **Netlify** através do plugin `@netlify/plugin-nextjs`, proporcionando tempos de resposta ultra-rápidos e cache automático para Server Components.

---

## 🤖 Ciclo de Desenvolvimento com IA Multi-Agentes

Este projeto utiliza um sistema de agentes autônomos e especializados integrados ao fluxo de desenvolvimento:

```mermaid
flowchart LR
    PM["@pm\n(write-specs)"] -->|Especificação Aprovada| ENG["@engeneer\n(generate-code)"]
    ENG -->|Código & Testes| QA["@qa\n(audit-code)"]
    QA -->|Auditoria de Qualidade| SEC["@security\n(audit-security)"]
    SEC -->|Auditoria de Segurança| DEVOPS["@devops\n(deploy-app)"]
    DEVOPS -->|Publicação em Produção| PROD[("🚀 Produção")]

    QA -.->|Bugs encontrados| ENG
    SEC -.->|Vulnerabilidades encontradas| ENG
    DEVOPS -.->|Falha de Deploy / Rollback| ENG
```

- **`@pm` (`write-specs`):** Converte ideias e requisitos em especificações detalhadas, objetivos, não-objetivos e critérios de aceite.
- **`@engeneer` (`generate-code`):** Implementa código limpo, tipado e com testes unitários cobrindo cenários de sucesso e erro.
- **`@qa` (`audit-code`):** Executa testes automatizados, validação de contratos OpenAPI e testes exploratórios de borda.
- **`@security` (`audit-security`):** Realiza modelagem de ameaças (STRIDE), auditoria de CVEs, verificação de segredos e conformidade com a LGPD.
- **`@devops` (`deploy-app`):** Cuida da integração contínua, infraestrutura como código, migrações seguras e procedimentos de rollback.

Para iniciar um novo ciclo completo de feature ou correção:
```
/startcycle <descrição da funcionalidade ou melhoria>
```

---

## 📄 Licença

Este projeto é distribuído sob a licença **MIT**. Consulte o arquivo [LICENSE](LICENSE) para mais detalhes.

---

<div align="center">
  <sub>Construído com excelência para a modernização da gestão pública municipal.</sub>
</div>
