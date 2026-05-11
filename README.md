# YesokVietnam

A backend service for the Yesok Vietnam project, built with **Go**, **Gin**, and **GORM**.

## Quick Start

```bash
# 1. Copy environment file and fill in your values
cp .env.example .env

# 2. Run
go run main.go
```

## Configuration

| Variable       | Default      | Description                         |
|----------------|-------------|-------------------------------------|
| `SERVER_PORT`  | `8080`      | HTTP server port                    |
| `DB_HOST`      | `localhost` | MySQL host                          |
| `DB_PORT`      | `3306`      | MySQL port                          |
| `DB_USER`      | `root`      | MySQL username                      |
| `DB_PASSWORD`  | —           | MySQL password                      |
| `DB_NAME`      | `yesok_vietnam` | MySQL database name             |
| `TG_BOT_TOKEN` | —           | Telegram bot token (required)       |

## API Endpoints

| Method | Path            | Auth | Description                     |
|--------|----------------|------|---------------------------------|
| GET    | `/health`       | No   | Health check                    |
| POST   | `/api/auth/tg`  | No   | Telegram initData login         |
| GET    | `/api/auth/me`  | Yes  | Get current user profile        |

## Telegram Login Flow

The frontend should call `window.Telegram.WebApp.initData` and send it to `POST /api/auth/tg`:

```json
{ "initData": "<the initData string>" }
```

On success, the server returns:

```json
{
  "token": "...",
  "user": { "id": 1, "tg_id": 12345, "username": "...", "role": "user", "balance": 0 },
  "is_new": false,
  "expire": 1700000000
}
```

Subsequent authenticated requests must include:

```
Authorization: Bearer <token>
```

## Project Structure

```
.
├── config/         Configuration
├── handlers/       HTTP handlers
├── middleware/      Auth middleware
├── models/         GORM models (User, Order)
├── pkg/telegram/   Telegram initData validator
├── html/           index.html (SPA entry)
├── static/         Static assets
└── main.go         Application entry point
```

## Data Models

### User
- `tg_id` — Telegram user ID (unique)
- `username`, `first_name`, `last_name`
- `role` — `admin | user | worker`
- `balance` — account balance (decimal)

### Order
- `order_no` — unique order number
- `user_id` — FK to User
- `status` — `pending | confirmed | in_progress | completed | cancelled | failed | refunded`
- `amount`, `currency`
- `tg_chat_id`, `tg_message_id` — Telegram references
- `worker_tg_id` — assigned worker
- `metadata` — JSON extra data
- `note` — admin note
