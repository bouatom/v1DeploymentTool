# V1 SG Deployment Tool

Containerized deployment console for scanning targets and running OS-aware installer workflows.

## Quick Start (Docker)

```bash
docker compose up --build
```

Web: `http://localhost:5173`  
API: `http://localhost:8080`

## Services

- **Postgres**: persistence for tasks, scans, deployments, credentials, and audit logs
- **API**: Go Fiber backend
- **Web**: React + Vite UI served by Nginx

## Data Persistence

`docker-compose.yml` uses named volumes:

- `postgres-data` for database storage
- `api-uploads` for uploaded installers

## Environment Variables

Set via `docker-compose.yml` or your shell.

### API

- `DATABASE_URL` (required)
- `APP_HTTP_ADDRESS` (default `:8080`)
- `CREDENTIALS_KEY` (required, base64)
- `CREDENTIALS_KEY_ID` (default `default`)
- `ADMIN_API_KEY` (required)
- `VIEWER_API_KEY` (optional)
- `RETENTION_DAYS` (default `90`)

### Web

- `VITE_API_BASE_URL` (default `http://localhost:8080`)
- `VITE_API_KEY` (optional)

## Installer Uploads

Uploaded binaries are stored under `/app/uploads` in the API container and served at:

```
http://<api-host>/uploads/<filename>
```

Targets must be able to reach this URL for downloads.

## Preflight Auth Check

Use `POST /api/preflight` to validate credentials and target reachability before deployment.

## Local Development (without Docker)

```bash
export DATABASE_URL="postgres://deploy:deploy@localhost:5432/deploy?sslmode=disable"
export CREDENTIALS_KEY="MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="
export CREDENTIALS_KEY_ID="dev"
export ADMIN_API_KEY="admin-dev-key"
export VIEWER_API_KEY="viewer-dev-key"
go run ./cmd/api
```

```bash
cd web
npm install
npm run dev
```
