# AGENTS.md - Agentic Coding Guidelines

This file provides guidance for agents working in this codebase.

## Project Overview

This is a Go-based multiplayer game server using RabbitMQ for message routing. The project follows standard Go conventions and uses the `amqp091-go` library for RabbitMQ communication.

## Build, Lint, and Test Commands

### Running Tests

```bash
# Run all tests
make test

# Run all tests with coverage
make test/cover

# Run a single test
go test -v -race -run TestName ./path/to/package
```

### Quality Control

```bash
# Run full audit (tidy check, format check, vet, staticcheck, vulnerability check)
make audit

# Check gofmt formatting
gofmt -l .

# Run go vet
go vet ./...

# Run staticcheck
go run honnef.co/go/tools/cmd/staticcheck@latest ./...
```

### Building and Running

```bash
# Build the application
make build

# Build and run
make run

# Build with live reload on file changes
make run/live
```

### Development Commands

```bash
# Tidy modules and format code
make tidy
```

## Code Style Guidelines

### Formatting

- Use `gofmt` or `go fmt` to format code automatically
- Run `make tidy` before committing (runs `go mod tidy`, `go fix`, `go fmt`)

### Imports

Group imports in the following order with blank lines between groups:

1. Standard library packages
2. External/third-party packages

```go
import (
    "context"
    "flag"
    "fmt"

    "github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
    "github.com/rabbitmq/amqp091-go"
)
```

### Naming Conventions

- **Variables**: camelCase (e.g., `client`, `config`, `gameState`)
- **Constants**: PascalCase for exported, camelCase for unexported (e.g., `RankInfantry`, `playingState`)
- **Types/Structs**: PascalCase (e.g., `GameState`, `Player`, `Unit`)
- **Functions**: PascalCase for exported, camelCase for unexported (e.g., `NewClient`, `getInput`)
- **Interfaces**: PascalCase, typically with `-er` suffix (e.g., `Reader`, `Publisher`)
- **Packages**: short, lowercase, no underscores (e.g., `gamelogic`, `rabbitmq`, `routing`)

### Error Handling

- Use `fmt.Errorf` with `%w` for wrapping errors:
  ```go
  return fmt.Errorf("failed to marshal json: %w", err)
  ```
- Return errors directly from constructors/factories
- Check errors immediately after calls that can fail
- Use sentinel errors for defined error conditions when appropriate

### Types

- Use the zero-value-friendly types when possible
- For config structs, define as a simple struct:
  ```go
  type config struct {
      addr string
  }
  ```
- Use specific types (not `interface{}` or `any`) when the type is known

### Context and Cancellation

- Always accept `context.Context` as the first parameter for functions that may block
- Use `context.WithTimeout` or `context.WithCancel` for operations with deadlines
- Use `signal.NotifyContext` for handling OS signals:
  ```go
  ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
  defer stop()
  ```

### Resource Management

- Use `defer` for cleanup (especially for closing connections):
  ```go
  client, err := rabbitmq.NewClient(cfg.addr)
  if err != nil {
      return err
  }
  defer client.Close()
  ```

### Concurrency

- Use channels for communication between goroutines
- Use `sync.Mutex` for protecting shared state
- Always check for context cancellation in loops:
  ```go
  for {
      select {
      case <-ctx.Done():
          return ctx.Err()
      default:
      }
      // work...
  }
  ```

### RabbitMQ Patterns

- Use `Durable` for queues that should survive broker restarts
- Use `Transient` for temporary queues
- Use topic exchanges (`ExchangePerilTopic`) for pattern-based routing
- Use direct exchanges (`ExchangePerilDirect`) for specific routing keys
- Use `PublishJSON` for JSON-serialized messages
- Use `PublishGob` for Gob-serialized messages (more efficient for Go-to-Go)

### Testing

- Create test files with `_test.go` suffix in the same package
- Use `testing.T` for test functions
- Test both success and failure cases
- Use table-driven tests when testing multiple input cases

## Project Structure

```
cmd/
  server/    - Server main package
  client/    - Client main package
internal/
  gamelogic/ - Game logic and state management
  rabbitmq/  - RabbitMQ client and publishing/subscribing
  routing/   - Exchange names, routing keys, and message models
```

## Common Patterns

### Flag Parsing

```go
flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
flags.SetOutput(stdout)
flags.StringVar(&cfg.addr, "addr", "default", "description")
if err := flags.Parse(args[1:]); err != nil {
    return err
}
```

### Graceful Shutdown

```go
func main() {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    if err := run(ctx, os.Args, os.Stdout); err != nil {
        fmt.Fprintf(os.Stderr, "%v\n", err)
        os.Exit(1)
    }

    log.Println("shutdown complete")
    os.Exit(0)
}
```

### Message Subscription Handler

```go
rabbitmq.SubscribeJSON(client, exchange, queue, key, durable, func(msg MyMessage) rabbitmq.AckType {
    // process message
    return rabbitmq.Ack
})
```

## Dependencies

- `github.com/rabbitmq/amqp091-go` - RabbitMQ client library
- Go 1.22.1

## Additional Notes

- The server binary is built to `/tmp/bin/server` by default
- The client binary is built similarly (set `binary_name` in Makefile)
- RabbitMQ must be running locally (default: `amqp://guest:guest@localhost:5672`)