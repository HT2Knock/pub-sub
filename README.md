# Peril - Multiplayer Game Server

A Go-based multiplayer game server using RabbitMQ for message routing. This is a Peril-style game where players can spawn units, move them across continents, and engage in battle.

## Prerequisites

- Go 1.22.1+
- RabbitMQ (running locally on `amqp://guest:guest@localhost:5672`)

## Quick Start

### Install Dependencies

```bash
go mod download
```

### Run Tests

```bash
make test
```

### Build

```bash
make build
```

### Run Server and Client

Start the server in one terminal:

```bash
make run
# or for live reload on changes
make run/live
```

In another terminal, run the client:

```bash
go run ./cmd/client
```

## Usage

### Server Commands

- `pause` - Pause the game
- `resume` - Resume the game
- `quit` - Exit the server
- `help` - Show help

### Client Commands

- `spawn <location> <rank>` - Spawn a unit (e.g., `spawn europe infantry`)
- `move <location> <unitID>...` - Move units (e.g., `move asia 1 2`)
- `status` - Show your units
- `spam <n>` - Send test messages
- `quit` - Exit the client
- `help` - Show help

### Locations

`americas`, `europe`, `africa`, `asia`, `australia`, `antarctica`

### Unit Ranks

`infantry`, `cavalry`, `artillery`

## Project Structure

```
cmd/
  server/      # Server main package
  client/      # Client main package
internal/
  gamelogic/  # Game logic and state management
  rabbitmq/   # RabbitMQ client and publishing/subscribing
  routing/    # Exchange names, routing keys, and message models
```

## Available Make Commands

- `make test` - Run all tests
- `make test/cover` - Run tests with coverage
- `make audit` - Run quality checks (tidy, vet, staticcheck, vuln scan)
- `make tidy` - Format and tidy code
- `make build` - Build the server binary
- `make run` - Build and run the server
- `make run/live` - Run with live reload

## License

MIT
