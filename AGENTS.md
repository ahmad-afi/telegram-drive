# AGENTS.md

## Project Overview

Telegram Drive — Google-Drive-style file storage using Telegram Saved Messages as the backend via MTProto API.

## Architecture

- **Backend** (`backend/`): Go + Fiber, wraps `gotd/td` MTProto. Stateless: no DB, no session store. Each request carries a Telegram session string via `X-Session` header. Server reconstructs a fresh Telegram client per request.
- **Frontend** (`frontend/`): React 19 + TypeScript + Vite SPA. Session stored in `localStorage`.
- **Storage**: Files live in user's own Saved Messages. Virtual folders encoded as `#dir:FolderName/filename` caption tags. No database needed — folders derived by scanning message captions.
- **Login**: QR code flow. `POST /api/qr/start` returns PNG QR (base64). User scans in Telegram mobile app. Frontend polls `GET /api/qr/poll/:id` until session string returned.

## Key Design Decisions

- **Stateless server**: No DB, no session store. Browser owns the session. Each API call reconstructs a gotd client from base64 session JSON.
- **Virtual folders via caption tags**: `#dir:Documents/file.pdf` → folder "Documents", file "file.pdf". Root files have bare filename as caption. `parseCaption` uses `LastIndex("/")` to support nested folder names.
- **No `POST /api/folders`**: Folders are implicit — created when a file is uploaded with a folder tag. Endpoint returns 501 intentionally.
- **Upload progress via XHR**: `upload.onprogress` gives browser→server progress. No SSE/WebSocket needed. Server→Telegram progress is not surfaced (blocking, 30min cap).

## Tech Stack

- Backend: Go 1.26, gofiber/fiber/v2, gotd/td (MTProto), joho/godotenv, rsc.io/qr
- Frontend: React 19, Vite 8, TypeScript 6, oxlint. No routing/state/UI libraries.
- Deploy: Docker (multi-stage builds), docker-compose

## Project Structure

```
backend/
  main.go                    # Entrypoint, Fiber app, routes
  internal/
    config/config.go         # Env loading (TG_APP_ID, TG_APP_HASH, PORT, CORS_ORIGIN)
    middleware/session.go    # X-Session header extraction
    handlers/
      qr.go                  # QR login start/poll
      file.go                # File CRUD (list, upload, download, delete)
      folder.go              # Folder list (create is 501)
    telegram/
      client.go              # MTProto wrapper: Run(), Upload, Download, Delete, ListFiles, ListFolders
      qrlogin.go             # QR login manager (in-memory map, 3min timeout)
      helpers.go             # tagCaption, parseCaption, scanHistory, searchFiles, parseMessage
      util.go                # encodeBase64
frontend/
  src/
    main.tsx                 # React entry
    App.tsx                  # Auth state, dark mode toggle
    Login.tsx                # QR login screen
    Drive.tsx                # File browser (sidebar, search, upload, download, delete, drag-drop, pagination)
    api.ts                   # API client (fetch + XHR for upload progress)
    fmt.ts                   # fmtSize helper
    fmt.test.ts              # Self-check test
    App.css                  # All styles (CSS variables, dark mode, responsive)
    index.css                # Global reset
```

## Commands

### Backend
```bash
cd backend
cp .env.example .env          # fill TG_APP_ID, TG_APP_HASH
go run .                      # start server on :8080
go test ./...                 # run tests
go build ./...                # compile check
```

### Frontend
```bash
cd frontend
cp .env.example .env          # VITE_API_URL=http://localhost:8080
npm install
npm run dev                   # dev server
npm run build                 # production build to dist/
npm run lint                  # oxlint
npx tsx --test src/fmt.test.ts  # run self-check
```

### Docker
```bash
cp .env.example .env          # fill TG_APP_ID, TG_APP_HASH, VITE_API_URL
docker compose up --build     # backend :8080, frontend :3000
```

## API

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/qr/start` | — | Start QR login, returns `{loginId, qrImage}` |
| GET | `/api/qr/poll/:id` | — | Poll QR login status |
| GET | `/api/files?folder=root&q=&offset=0&limit=50` | X-Session | List/search files |
| POST | `/api/files` | X-Session | Upload (multipart: `file`, `folder`) |
| GET | `/api/files/:msgId/download` | X-Session | Download file |
| DELETE | `/api/files/:msgId` | X-Session | Delete file |
| GET | `/api/folders` | X-Session | List virtual folders |
| POST | `/api/folders` | X-Session | 501 — folders are implicit |

## Environment Variables

| Var | Required | Default | Description |
|-----|----------|---------|-------------|
| `TG_APP_ID` | yes | — | Telegram app ID from my.telegram.org |
| `TG_APP_HASH` | yes | — | Telegram app hash |
| `PORT` | no | `8080` | Backend port |
| `CORS_ORIGIN` | no | `*` | Comma-separated allowed origins |
| `VITE_API_URL` | no | `http://localhost:8080` | Backend URL for frontend |

## Completed Phases

1. **Polish & Testing** — backend tests (caption tag/parse round-trip, session middleware), frontend self-check (`fmtSize`), fixed `parseCaption` nested-folder bug, root README, index.html title.
2. **Search & Pagination** — `searchFiles` respects folder context, `ListFiles` supports `offset`/`limit`, "Muat lebih banyak" pagination UI.
3. **Upload Progress** — XHR `upload.onprogress`, real-time progress bar, zero backend changes.
4. **Production Deploy** — Dockerfiles (multi-stage), `docker-compose.yml`, configurable CORS origin.
5. **UI/UX** — Dark mode (auto-detect + toggle + localStorage), drag-and-drop upload, responsive mobile layout.

## Known Limitations (ponytail)

- `backend/internal/handlers/file.go` — blocking upload tied to request lifetime, 30min cap. Move to job queue when files routinely exceed ~1GB.
- `backend/internal/handlers/folder.go` — `POST /api/folders` is 501. Pre-created empty folders need a placeholder message with just the tag caption.
- Session strings in `localStorage` (XSS-exposable). Trade-off of stateless design. Upgrade path: HttpOnly cookies + server-side session store.
- `searchFiles` hard-capped at 100 results from Telegram's `MessagesSearch` API.
