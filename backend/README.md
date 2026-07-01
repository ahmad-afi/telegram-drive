# Telegram Drive — Backend

Go + Fiber wrapper for Telegram MTProto (gotd/td). **Stateless**: the server stores nothing — each request carries the user's session string in an `X-Session` header. Files live in the user's own Saved Messages.
## Prerequisites

- Go 1.26+
- A Telegram API app: get `api_id` + `api_hash` from https://my.telegram.org

## Getting Telegram API Credentials

1. Buka **https://my.telegram.org** dan login dengan nomor HP Telegram kamu.
2. Klik **API development tools**.
3. Isi **App title** dan **Short name** (bebas, misal `Telegram Drive`).
4. Pilih platform **Other**, klik **Create application**.
5. Salin **App api_id** (angka) → taruh di `.env` sebagai `TG_APP_ID`.
6. Salin **App api_hash** (string) → taruh di `.env` sebagai `TG_APP_HASH`.

## Environment Variables

| Var | Required | Default | Description |
|-----|----------|---------|-------------|
| `TG_APP_ID` | yes | — | Telegram App ID dari my.telegram.org |
| `TG_APP_HASH` | yes | — | Telegram App Hash dari my.telegram.org |
| `PORT` | no | `8080` | Port backend |
| `CORS_ORIGIN` | no | `*` | Comma-separated allowed origins (misal `http://localhost:3000,https://app.example.com`) |

## Setup

```sh
cp .env.example .env   # isi TG_APP_ID dan TG_APP_HASH dari my.telegram.org
go mod tidy
```

## Run

```sh
go run .
```

Server listens on `:8080`.

## How login works

1. `POST /api/qr/start` — server starts a QR login flow, returns `{loginId, qrImage}` (PNG base64).
2. User scans the QR with their Telegram app (Settings → Devices → Link Device).
3. `GET /api/qr/poll/:id` — frontend polls until `{status: "ok", session, name}`.
4. The `session` string is stored in the browser's `localStorage`.
5. All subsequent requests send `X-Session: <base64>` — the server uses it per-request and discards it.

The server **never persists** session strings. Each user's files live in their own Telegram account.

## How folders work

Folders are virtual, stored as caption tags on Saved Messages messages:
- A file in "Documents" → caption `#dir:Documents/myfile.pdf`
- Root files → caption is just the filename

`GET /folders` scans Saved Messages and returns unique folder names. `GET /files?folder=Documents` returns files tagged with that folder. No database needed.

## API

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/api/qr/start` | — | Start QR login, returns QR image |
| GET | `/api/qr/poll/:id` | — | Poll QR login status |
| GET | `/api/files?folder=root` | X-Session | List files in folder |
| GET | `/api/files?q=term` | X-Session | Search files by name |
| POST | `/api/files` | X-Session | Upload (multipart `file` + `folder`) |
| GET | `/api/files/:msgId/download` | X-Session | Stream file |
| DELETE | `/api/files/:msgId` | X-Session | Delete file |
| GET | `/api/folders` | X-Session | List virtual folders |
