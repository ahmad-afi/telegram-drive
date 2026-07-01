# Telegram Drive

Google-Drive-style file storage using Telegram Saved Messages as the backend.

## Architecture

- **Backend** — Go + Fiber, wraps `gotd/td` MTProto. Stateless: no DB, no session store. Each request carries a Telegram session string via `X-Session` header.
- **Frontend** — React 19 + TypeScript + Vite SPA. Session stored in `localStorage`.
- **Storage** — Files live in the user's own Saved Messages. Virtual folders are encoded as `#dir:FolderName/filename` caption tags.

## Quick Start

### Backend

```bash
cd backend
cp .env.example .env   # fill in TG_APP_ID and TG_APP_HASH
go run .
```

Get `TG_APP_ID` and `TG_APP_HASH` from https://my.telegram.org.

### Frontend

```bash
cd frontend
cp .env.example .env   # defaults to http://localhost:8080
npm install
npm run dev
```

## API

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/qr/start` | — | Start QR login, returns `{loginId, qrImage}` |
| GET | `/api/qr/poll/:id` | — | Poll QR login status |
| GET | `/api/files?folder=root` | X-Session | List files in folder |
| GET | `/api/files?q=term` | X-Session | Search files by name |
| POST | `/api/files` | X-Session | Upload file (multipart: `file`, `folder`) |
| GET | `/api/files/:msgId/download` | X-Session | Download file |
| DELETE | `/api/files/:msgId` | X-Session | Delete file |
| GET | `/api/folders` | X-Session | List virtual folders |

## Security Note

Session strings are stored in `localStorage` (XSS-exposable). This is a deliberate trade-off of the stateless design. For production use, consider HttpOnly cookies with a server-side session store.

## Docker Deploy

```bash
cp .env.example .env   # fill in TG_APP_ID, TG_APP_HASH, VITE_API_URL
docker compose up --build
```

Backend on `:8080`, frontend on `:3000`. Set `VITE_API_URL` to the public backend URL for production.
