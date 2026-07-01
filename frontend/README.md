# Telegram Drive — Frontend

React + TypeScript + Vite UI for the Telegram Drive backend.

## Setup

```sh
cp .env.example .env   # point VITE_API_URL at your backend
npm install
npm run dev
```

## How it works

1. **Login**: Click "Mulai Login" → QR code appears → scan with Telegram app → done.
2. **Session** is stored in `localStorage` (key: `tg_session`). The server never stores it.
3. **Drive**: Browse files in your Saved Messages. Folders are virtual (caption tags).
4. **Upload**: Files go to Saved Messages with a `#dir:` caption tag for folder routing.
5. **Download**: Files are fetched as blobs and downloaded via anchor links.

## Build

```sh
npm run build
```
