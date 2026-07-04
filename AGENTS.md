# AGENTS.md

## Project Overview

Microphoto is a distributed image/video processing platform. A Go backend splits images into fragments, processes them in parallel across worker nodes, and reconstructs the final result. The frontend is an Astro 7 app with React islands. All UI text is **Spanish**.

Three backend services: **Coordinator** (port 8080), **Worker**, **Reaper**. Infrastructure: Redis (queue/pub-sub), MinIO (object storage).

## Commands

### Frontend (`frontend/`)

```bash
bun install        # install deps
bun dev            # dev server (Astro)
bun run build      # production build
bun run preview    # preview production build
bunx biome check . # lint + format check
bunx biome check --write .  # auto-fix
```

### Backend (`backend/`)

```bash
go build ./cmd/coordinator  # build coordinator
go build ./cmd/worker       # build worker (requires libvips + CGO_ENABLED=1)
go build ./cmd/reaper       # build reaper
just build                  # build all three binaries (requires just)
just docker                 # build Docker image
just docker-all             # build all Docker images
just proto                  # regenerate protobuf code
```

Lint: `golangci-lint run` (uses `.golangci.yml`; errcheck disabled)

### Infrastructure

```bash
docker compose up           # start Redis, MinIO, all 3 services
```

## Key Constraints

- **Worker binary requires CGO** (`CGO_ENABLED=1`) because of `bimg`/libvips. Coordinator and Reaper build with CGO disabled.
- **Nix dev shell** (`flake.nix` + direnv) provides the full toolchain. GOPATH is remapped to `.go/` inside the project.
- **Biome** is the linter/formatter for the frontend. Enforces single quotes, semicolons, `@/` absolute imports (relative imports are banned).
- **Node >=22.12.0** required for frontend.
- **Max upload size**: 2GB (configurable via env).
- Frontend connects to backend via `PUBLIC_API_URL` env var (defaults `http://localhost:8080`).

## Architecture

```
backend/
  cmd/coordinator/   # HTTP API, SSE, file upload, preview
  cmd/worker/        # Queue consumer, image/video processing
  cmd/reaper/        # Timeout detection, job rescheduling
  pkg/               # Shared packages (Redis, MinIO, metrics clients)
  proto/             # Protobuf definitions (buf.yaml, jobs.proto)

frontend/
  src/components/    # React components (App, ImageUploader, ImageEditor, ProgressTracker, etc.)
  src/hooks/         # useSSE, useImagePreview, useTaskHistory
  src/lib/api.ts     # API client (upload, preview, SSE, results)
  src/pages/         # Astro pages (/ and /app)
  src/styles/        # global.css (oklch color system, dark mode only)

kube/                # Helm charts, SOPS secrets, Helmfile
```

### Data Flow

Upload → Coordinator saves to MinIO, pushes SLICE job to Redis → Worker splits into fragments → Workers process in parallel → Last worker triggers RECONSTRUCT → Final image uploaded to MinIO → Frontend receives completion via SSE.

## Conventions

- Path aliases: `@/*` maps to `./src/*` (frontend TypeScript).
- `verbatimModuleSyntax: true` in tsconfig.
- shadcn/ui with `radix-luma` style, Tabler icons.
- Tailwind CSS v4 with Vite plugin.
- Protobuf: buf v2 toolchain, codegen via `just proto`.
- All commit messages and code comments are in Spanish.
- No test suites exist currently.
- No CI/CD pipelines configured.
