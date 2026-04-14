# AgentShield Frontend

React + Vite dashboard client for AgentShield.

## Features

- Supabase login (Google OAuth)
- Dashboard metrics from live scan/report data
- Scan creation + start workflow
- Scan monitoring with polling + WebSocket status feed
- Report comparison
- Judge calibration trigger/report view
- Settings/diagnostics checks (API + auth health)

## Environment

Copy env example and fill values:

```bash
cd frontend
cp .env.example .env
```

Required:

- `VITE_SUPABASE_URL`
- `VITE_SUPABASE_ANON_KEY`
- `VITE_API_URL` (default local gateway: `http://localhost:8080`)

## Run

```bash
cd frontend
npm install
npm run dev
```

Default dev URL: `http://localhost:3000`

## Build

```bash
cd frontend
npm run lint
npm run build
```

## API Compatibility

This frontend expects the gateway API prefix `/api/v1` and WebSocket route `/ws/scans/:id/status`.
