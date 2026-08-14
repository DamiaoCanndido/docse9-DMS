# 📘 Guia de Deploy & CI/CD — GCP Cloud Run + Neon/Supabase

Este guia detalha o passo a passo para configurar a infraestrutura de nuvem, banco de dados e pipeline de CI/CD para o backend do **Docseq DMS**.

---

## 🏗️ 1. Arquitetura da Solução

```mermaid
flowchart LR
    A[GitHub Repo] -->|Push na master| B[GitHub Actions CI/CD]
    B -->|Build & Push| C[GCP Artifact Registry]
    B -->|Deploy| D[GCP Cloud Run - Go API]
    D -->|Segredos injetados| E[GCP Secret Manager]
    D -->|Queries SQL via TLS| F[(Neon / Supabase PostgreSQL)]
```

- **Backend:** Google Cloud Run (Container serverless, escala até zero quando ocioso, cold start < 1s).
- **Banco de Dados:** PostgreSQL hospedado no Neon.tech (Tier Free com 0.5 GB de storage) ou Supabase (500 MB).
- **Segredos:** GCP Secret Manager (protege `DATABASE_URL`, `JWT_SECRET`, `DEFAULT_ADMIN_PASSWORD`).
- **CI/CD:** GitHub Actions com testes automatizados e deploy contínuo.

---

## 🐘 2. Passo 1: Criar o Banco de Dados (Neon ou Supabase)

### Opção A: Neon.tech (Recomendado para Serverless)
1. Crie uma conta gratuita em [neon.tech](https://neon.tech).
2. Crie um novo projeto chamado `docseq-dms`.
3. Escolha uma região próxima (ex: `us-east-1` ou `sa-east-1` se disponível).
4. Copie a **Connection String** no formato pooling (`postgres://...`). Exemplo:
   ```text
   postgres://username:password@ep-cold-pond-123456-pooler.us-east-2.aws.neon.tech/neondb?sslmode=require
   ```

### Opção B: Supabase
1. Crie uma conta em [supabase.com](https://supabase.com).
2. Crie um novo projeto `docseq-dms` na região `São Paulo (sa-east-1)`.
3. Em **Project Settings -> Database**, copie a **URI** de conexão no modo *Transaction Pooler* (porta 6543) ou *Session* (porta 5432) com `sslmode=require`.

---

## ☁️ 3. Passo 2: Configurar o Google Cloud Platform (GCP)

### 3.1. Criar Projeto e Habilitar APIs
No console do GCP ou via terminal `gcloud`:
```bash
# Defina o ID do projeto
export PROJECT_ID="docseq-dms-prod"
export REGION="southamerica-east1" # ou us-east4

# Habilite as APIs necessárias
gcloud services enable \
  run.googleapis.com \
  artifactregistry.googleapis.com \
  secretmanager.googleapis.com \
  iam.googleapis.com \
  --project $PROJECT_ID
```

### 3.2. Criar Repositório no Artifact Registry
```bash
gcloud artifacts repositories create docseq-artifacts \
  --repository-format=docker \
  --location=$REGION \
  --description="Docker repository for Docseq DMS" \
  --project $PROJECT_ID
```

### 3.3. Criar os Segredos no Secret Manager
```bash
# 1. Connection string do PostgreSQL
echo -n "postgres://user:pass@host/db?sslmode=require" | \
  gcloud secrets create DOCSEQ_DATABASE_URL --data-file=- --project $PROJECT_ID

# 2. JWT Secret (gere uma chave forte de 32+ caracteres)
openssl rand -base64 32 | tr -d '\n' | \
  gcloud secrets create DOCSEQ_JWT_SECRET --data-file=- --project $PROJECT_ID

# 3. Senha inicial do Administrador padrão
echo -n "SuaSenhaAdminSegura123!" | \
  gcloud secrets create DOCSEQ_ADMIN_PASSWORD --data-file=- --project $PROJECT_ID
```

### 3.4. Criar Service Account para o GitHub Actions (Menor Privilégio)
```bash
# 1. Criar a Service Account do CI/CD
gcloud iam service-accounts create github-actions-deployer \
  --display-name="GitHub Actions Deployer" \
  --project $PROJECT_ID

export SA_EMAIL="github-actions-deployer@${PROJECT_ID}.iam.gserviceaccount.com"

# 2. Conceder permissões para Cloud Run e Artifact Registry
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/run.admin"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/artifactregistry.writer"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/iam.serviceAccountUser"

# 3. Conceder permissão para o Cloud Run ler os segredos no Secret Manager
# (A conta padrão de runtime do Cloud Run precisa acessar os segredos)
export COMPUTE_SA="$(gcloud projects describe $PROJECT_ID --format='value(projectNumber)')-compute@developer.gserviceaccount.com"

gcloud secrets add-iam-policy-binding DOCSEQ_DATABASE_URL \
  --member="serviceAccount:${COMPUTE_SA}" \
  --role="roles/secretmanager.secretAccessor" \
  --project $PROJECT_ID

gcloud secrets add-iam-policy-binding DOCSEQ_JWT_SECRET \
  --member="serviceAccount:${COMPUTE_SA}" \
  --role="roles/secretmanager.secretAccessor" \
  --project $PROJECT_ID

gcloud secrets add-iam-policy-binding DOCSEQ_ADMIN_PASSWORD \
  --member="serviceAccount:${COMPUTE_SA}" \
  --role="roles/secretmanager.secretAccessor" \
  --project $PROJECT_ID

# 4. Gerar chave JSON para colocar no GitHub Secrets
gcloud iam service-accounts keys create gcp-key.json \
  --iam-account="${SA_EMAIL}" \
  --project $PROJECT_ID
```

> [!CAUTION]
> Guarde o arquivo `gcp-key.json` com segurança e **nunca** o commite no repositório Git.

---

## 🔒 4. Passo 3: Configurar o GitHub Actions (Secrets & Variables)

Acesse o repositório no GitHub: **Settings -> Secrets and variables -> Actions**.

### 🔑 Repository Secrets
| Nome do Secret | Valor |
| :--- | :--- |
| `GCP_SA_KEY` | Conteúdo JSON completo do arquivo `gcp-key.json` gerado no passo 3.4 |

### 🏷️ Repository Variables
| Nome da Variável | Exemplo de Valor |
| :--- | :--- |
| `GCP_PROJECT_ID` | `docseq-dms-prod` |
| `GCP_REGION` | `southamerica-east1` (ou `us-east4`) |
| `GAR_REPOSITORY` | `docseq-artifacts` |
| `CLOUD_RUN_SERVICE`| `docseq-backend-api` |

---

## 🚀 5. Executando o Primeiro Deploy

1. Faça o commit e push das alterações para a branch `master`:
   ```bash
   git add .
   git commit -m "feat(infra): setup Dockerfile, CI/CD and GCP Cloud Run deployment"
   git push origin master
   ```
2. Acompanhe a execução em **Actions** no GitHub:
   - A pipeline executará os testes unitários e validações de código.
   - Fará o build da imagem Docker otimizada.
   - Publicará no Google Artifact Registry.
   - Fará o deploy no Google Cloud Run e validará o endpoint `/health`.
3. Ao finalizar, o link HTTPS do Cloud Run estará visível nos logs do GitHub Actions.

---

## 🔄 6. Rollback e Procedimentos Operacionais (@devops)

### Rollback Imediato via Console GCP ou CLI
Se uma versão apresentar instabilidade em produção:
```bash
# Listar as revisões anteriores
gcloud run revisions list --service=docseq-backend-api --region=$REGION

# Reverter o tráfego 100% para a revisão estável anterior instantaneamente:
gcloud run services update-traffic docseq-backend-api \
  --region=$REGION \
  --to-revisions=docseq-backend-api-00001-abc=100
```
