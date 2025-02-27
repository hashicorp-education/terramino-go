# Terramino Go

A HashiCorp-themed Tetris-like game with web and CLI interfaces, built in Go.

## Quick Start

```bash
# Start all services
docker compose up

# Play in browser
open http://localhost:8081

# Play in terminal
docker compose exec -it backend ./terramino-cli
```

## Development

Prerequisites: Docker Compose, Go 1.22+

Environment variables (`.env`):

- `REDIS_HOST`: Redis hostname (default: redis)
- `REDIS_PORT`: Redis port (default: 6379)
- `TERRAMINO_PORT`: Backend port (default: 8080)

Local development:

```bash
# Run backend
go run main.go

# Run CLI
go run cmd/cli/main.go
```

## Project Structure
```
├── cmd/cli/     # CLI client
├── internal/    # Game logic & high scores
├── web/        # Frontend assets
└── *.go        # Backend server
``` 