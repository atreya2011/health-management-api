# Health Management API

A backend system for a health management application following Domain-Driven Design (DDD) principles, built with Go, PostgreSQL, and Connect-RPC.

## Features

- User authentication via JWT
- Body composition tracking (weight, body fat percentage)
- Exercise records management
- Personal diary entries
- Health-related articles/columns

## Tech Stack

- **Language**: Go 1.21+
- **Database**: PostgreSQL 15
- **API**: Connect-RPC (gRPC-compatible HTTP API)
- **Authentication**: JWT
- **Code Generation**:
  - sqlc for type-safe SQL
  - buf for Protocol Buffers and Connect-RPC
- **CI/CD**: GitHub Actions with release-please

## Project Structure

The project follows a clean architecture approach with DDD principles:

```terminal
.
├── api/proto/                # Protocol Buffer definitions
├── cmd/server/               # Application entry point
├── configs/                  # Configuration files
├── db/
│   ├── migrations/           # SQL migrations
│   └── queries/              # SQL queries for sqlc
├── internal/
│   ├── application/          # Application services
│   ├── domain/               # Domain models and repository interfaces
│   └── infrastructure/       # Implementation details
│       ├── auth/             # JWT authentication
│       ├── config/           # Configuration loading
│       ├── log/              # Logging setup
│       ├── persistence/      # Repository implementations
│       └── rpc/              # Connect-RPC handlers
└── third_party/              # Generated OpenAPI specs
```

## Getting Started

### Prerequisites

- Go 1.21+
- Docker and Docker Compose
- Make

### Setup

1. Clone the repository:

   ```bash
   git clone https://github.com/atreya2011/health-management-api.git
   cd health-management-api
   ```

2. Install required tools:

   ```bash
   make setup-tools
   ```

3. Start the PostgreSQL database:

   ```bash
   make db-start
   ```

4. Run database migrations:

   ```bash
   make migrate-up
   ```

5. Generate code:

   ```bash
   make generate-all
   ```

6. Configure JWT:
   - Update `configs/config.yaml` with your JWT secret key

7. Build and run the server:

   ```bash
   make build
   ./bin/server serve
   ```

   Or use the make command:

   ```bash
   make run
   ```

## Development

### CLI Commands

The application provides a command-line interface with the following commands:

- `serve`: Start the API server

  ```bash
  ./bin/server serve [flags]
  ```

  Flags:
  - `-p, --port string`: Port to run the server on (overrides config)
  - `-v, --verbose`: Enable verbose output
  - `--config-path string`: Path to config directory (default "./configs")

- `seed`: Seed the database with mock data

  ```bash
  ./bin/server seed [flags]
  ```

  Flags:
  - `-d, --days int`: Number of days to generate mock data for (default 30)
  - `-v, --verbose`: Enable verbose output
  - `--config-path string`: Path to config directory (default "./configs")

### Common Make Commands

- `make help`: Display available commands
- `make build`: Build the server binary
- `make run`: Build and run the server
- `make seed`: Seed the database with mock data
- `make test`: Run tests
- `make test-integration`: Run integration tests with real database
- `make proto`: Generate Connect-RPC code from proto files
- `make sqlc`: Generate Go code from SQL queries
- `make migrate-up`: Apply database migrations
- `make migrate-down`: Revert the last database migration

### Adding New Features

1. Define the domain model in `internal/domain/`
2. Create repository interface in `internal/domain/`
3. Implement SQL queries in `db/queries/`
4. Implement repository in `internal/infrastructure/persistence/postgres/`
5. Create application service in `internal/application/`
6. Define API in Protocol Buffers (`api/proto/`)
7. Implement Connect-RPC handler in `internal/infrastructure/rpc/handlers/`
8. Register the handler in `cmd/server/cmd/serve.go`

## Testing

The project uses a combination of unit tests and integration tests:

### Unit Tests

Unit tests use mocks to test components in isolation. These tests are fast and don't require external dependencies.

```bash
make test
```

### Integration Tests

Integration tests use a real PostgreSQL database running in a Docker container. These tests provide more realistic testing of database interactions and are particularly useful for testing repositories and services.

```bash
make test-integration
```

or manually:

```bash
RUN_INTEGRATION_TESTS=true go test ./...
```

### Test Utilities

The `internal/testutil` package provides utilities for testing with a real database:

- Spinning up a PostgreSQL Docker container
- Running migrations to set up the schema
- Creating test data
- Cleaning up after tests

See `internal/testutil/README.md` for more details.
