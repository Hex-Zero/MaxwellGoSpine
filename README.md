# MaxwellGoSpine

[![CI](https://github.com/Hex-Zero/MaxwellGoSpine/actions/workflows/ci.yml/badge.svg)](https://github.com/Hex-Zero/MaxwellGoSpine/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/hex-zero/MaxwellGoSpine)](https://goreportcard.com/report/github.com/hex-zero/MaxwellGoSpine)
<!-- TODO: Add coverage badge (Codecov or Coveralls) after uploading reports there -->

Production-ready Go 1.22 REST API spine (net/http + chi) with layered architecture, PostgreSQL, structured logging, metrics, health/readiness, graceful shutdown.

## Quick Start

```bash
cp .env.example .env # edit values
docker compose up -d db
go mod tidy
go run ./cmd/server

# or with docker-compose (app + db)
docker compose up --build
```

## Make Targets

```bash
make run    # run dev server
make test   # run tests
make lint   # golangci-lint
make build  # build binary
make migrate-up   # apply migrations (requires migrate CLI)
make auto-commit  # run lint/test/build then auto commit & push if changes
make hooks-install # configure local git to use .githooks (pre-commit)
```

## Environment Variables

| Name | Required | Default | Description |
|------|----------|---------|-------------|
| APP_NAME | no | maxwell-api | App name |
| ENV | no | dev | dev or prod |
| HTTP_PORT | no | 8080 | HTTP listen port |
| DB_DSN | yes | - | Postgres DSN |
| READ_TIMEOUT | no | 10s | Read timeout |
| WRITE_TIMEOUT | no | 15s | Write timeout |
| CORS_ORIGINS | no | (empty) | Comma list of allowed origins |
| LOG_LEVEL | no | info | zap log level |
| PPROF_ENABLED | no | 0 | Enable /debug/pprof when 1 |
| CACHE_MAX_COST | no | 10000 | Ristretto max cost (approx entries) |
| CACHE_NUM_COUNTERS | no | 100000 | Ristretto counters (10x max items) |
| CACHE_BUFFER_ITEMS | no | 64 | Ristretto buffer items |
| REDIS_ADDR | no | (empty) | Redis host:port to enable shared cache |
| REDIS_PASSWORD | no | (empty) | Redis password |
| REDIS_DB | no | 0 | Redis DB number |

## API Examples

```bash
curl -X POST http://localhost:8080/v1/users -d '{"name":"Alice","email":"alice@example.com"}' -H 'Content-Type: application/json'
curl http://localhost:8080/v1/users?page=1&page_size=10
curl http://localhost:8080/v1/users/{uuid}
curl -X PATCH http://localhost:8080/v1/users/{uuid} -d '{"name":"New"}' -H 'Content-Type: application/json'
curl -X DELETE http://localhost:8080/v1/users/{uuid}
curl http://localhost:8080/metrics
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
```

## Tests & Lint

```bash
make test
make lint
```

## Migrations

Uses plain SQL compatible with golang-migrate.

```bash
migrate -path migrations -database $DB_DSN up
```

## OpenAPI

Minimal `openapi.yaml` included; view interactive docs at `/docs` (ReDoc) or raw spec at `/openapi.yaml`.

## Seed Data

Example seed script at `scripts/seed.sql`:

```bash
psql "$DB_DSN" -f scripts/seed.sql
```

## Notes

* Layers: cmd/server, internal/* (config, log, middleware, http handlers/render, core domain/services, storage).
* Replace module path in `go.mod` with your repository path if different.
* No global mutable singletons; dependencies passed via constructors.
* Enhancements: soft deletes, email normalization, merge-patch updates.
* New make target `auto-commit` to run checks then commit & push changes.
* Caching: layered Ristretto (in-process) + optional Redis; ETag middleware for GET responses.
* Pre-commit hook: run `make hooks-install` once to enable automatic gofmt + golangci-lint checks before each commit.

## Deployment (Three Image Build Paths)

You can produce & publish the container image via any (or all) of these options:

### 1. Local Docker (manual)

1. Install Docker Desktop (Windows/Mac) or Docker Engine (Linux).
2. Authenticate to ECR:

```powershell
$acct = (aws sts get-caller-identity --query Account --output text --profile maxwell)
$repoUri = "$acct.dkr.ecr.us-east-1.amazonaws.com/maxwell-api"
aws ecr get-login-password --region us-east-1 --profile maxwell | docker login --username AWS --password-stdin $repoUri
docker build -t maxwell-api:latest .
docker tag maxwell-api:latest $repoUri:latest
docker push $repoUri:latest
```

1. Set `api_image` in `terraform.tfvars` to the pushed URI then `terraform apply`.

### 2. GitHub Actions (OIDC, no local creds in secrets)

1. Set Terraform variable `github_repo = "Hex-Zero/MaxwellGoSpine"` then apply to create an IAM role & (if needed) the GitHub OIDC provider.
2. Copy output `github_actions_oidc_role_arn` into a new GitHub Actions repository secret named `AWS_GITHUB_OIDC_ROLE_ARN`.
3. Push to `main` – workflow `.github/workflows/ecr-build.yml` builds & pushes `:latest` to ECR.
4. (Optional) Add Terraform `depends_on` or run `terraform apply -refresh-only` to pick up new image digests later if using immutable tags.

### 3. AWS CodeBuild (managed build in AWS)

1. Leave `enable_codebuild = true` (default). Terraform creates a CodeBuild project + IAM role (or supply existing role ARN via `codebuild_service_role_arn`).
2. Trigger a build in the AWS Console (or add a webhook) to build & push `latest` image.
3. Output logs appear in CloudWatch group `/codebuild/maxwell`.

### Terraform Deploy

1. Copy `terraform/terraform.tfvars.example` to `terraform/terraform.tfvars` and fill: `api_image`, `api_keys`, `db_dsn`, optional `github_repo`.
2. Install Terraform >= 1.6.
3. Run:

```powershell
terraform -chdir=terraform init
terraform -chdir=terraform plan -out tf.plan
terraform -chdir=terraform apply tf.plan
```

1. Note outputs: `alb_dns_name`, `api_url`. Use header `X-API-Key: <key>` for `/v1/*`.

### Updating the Service

After pushing a new image (any path), run `terraform apply` again if using an immutable tag or update the tag reference.

### HTTPS (Next Step)

Add ACM certificate + HTTPS listener (port 443) and redirect from 80 for production hardening.

