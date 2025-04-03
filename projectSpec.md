# Health Management Backend - Detailed Implementation Plan (Go, PostgreSQL, Connect-RPC)

**Objective:** Create the backend system for a health management application following DDD principles, using Go, PostgreSQL, Connect-RPC, sqlc, Buf, Auth0, and release-please.

**Target Audience:** LLM for code generation, Backend Developers.

**Assumptions:**
* Go 1.21+ is used (for `slog`).
* Auth0 is configured with an API (defining audience) and potentially rules for custom claims if needed.
* Developer environment has Go, Docker (for PostgreSQL & testing), `make`, `git`, `curl` installed.

---

**1. Project Setup & Tooling**

1. **Initialize Project:**
    * Create project directory: `mkdir healthapp-backend && cd healthapp-backend`
    * Initialize Go Module: `go mod init github.com/<your-org>/healthapp-backend` (Replace `<your-org>`)
    * Initialize Git: `git init`

2. **Directory Structure (Strict):** Create the following directories:

    ```
    .
    ├── api/
    │   └── proto/
    │       └── healthapp/
    │           └── v1/          # Proto definitions for API v1
    ├── cmd/
    │   └── server/              # Main application entry point (main.go)
    ├── configs/                 # Configuration files (e.g., config.yaml, .env)
    ├── db/
    │   ├── migrations/          # SQL migration files (golang-migrate format)
    │   └── queries/             # SQL query files for sqlc
    ├── internal/
    │   ├── application/         # Application services/use cases (e.g., user_service.go)
    │   ├── domain/              # Core domain entities (e.g., user.go) & repository interfaces (e.g., user_repository.go)
    │   └── infrastructure/
    │       ├── auth/            # Auth0 JWT verification middleware/interceptor (auth_interceptor.go)
    │       ├── config/          # Configuration loading logic (config.go)
    │       ├── log/             # Logging setup (logger.go)
    │       ├── persistence/     # Data persistence layer
    │       │   └── postgres/    # PostgreSQL repository implementations (e.g., user_repo_pg.go)
    │       │       └── db/      # sqlc generated code (!! AUTO-GENERATED !!)
    │       └── rpc/             # Connect-RPC implementation
    │           ├── gen/         # Connect-RPC generated code (!! AUTO-GENERATED !!)
    │           └── handlers/    # Service handler implementations (e.g., user_handler.go)
    ├── third_party/
    │   ├── openapi/             # Generated OpenAPI v2 specification (e.g., healthapp/v1/api.swagger.json)
    │   └── swagger-ui/          # (Optional) Swagger UI static files for viewing OpenAPI spec
    ├── .github/
    │   └── workflows/           # GitHub Actions workflows (ci.yml, release-please.yml)
    ├── .gitignore
    ├── buf.gen.yaml             # Buf code generation configuration
    ├── buf.yaml                 # Buf linting/breaking change configuration
    ├── docker-compose.yml       # For local development PostgreSQL
    ├── go.mod
    ├── go.sum
    ├── Makefile                 # Make commands for common tasks
    └── sqlc.yaml                # sqlc configuration
    ```

3. **Install Tools (Go Binaries):**

    ```bash
    go install github.com/bufbuild/buf/cmd/buf@latest
    go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest # For OpenAPI
    go install github.com/google/wire/cmd/wire@latest # Optional: For dependency injection
    ```

    *Verify installations (`buf --version`, `sqlc version`, etc.)*

4. **Core Go Dependencies:**

    ```bash
    go get connectrpc.com/connect              # Connect-RPC core
    go get google.golang.org/protobuf/proto # Protobuf Go runtime
    go get github.com/jackc/pgx/v5/pgxpool     # PostgreSQL driver
    go get github.com/golang-jwt/jwt/v5        # JWT parsing/validation
    go get github.com/MicahParks/keyfunc/v2    # JWKS key fetching/caching (for JWT validation)
    go get github.com/spf13/viper              # Configuration management
    go get github.com/google/uuid              # UUID generation
    # Optional (for testing/mocks)
    go get github.com/stretchr/testify/assert
    go get github.com/stretchr/testify/require
    go get github.com/stretchr/testify/mock
    ```

5. **Configuration (`buf.yaml`):** Create `buf.yaml`

    ```yaml
    version: v1
    name: buf.build/<your-org>/healthapp-backend
    deps:
      - buf.build/googleapis/googleapis
    lint:
      use:
        - DEFAULT
      except:
        - RPC_REQUEST_STANDARD_NAME # Optional: relax if needed
        - RPC_RESPONSE_STANDARD_NAME # Optional: relax if needed
    breaking:
      use:
        - FILE # Detect breaking changes against previous versions
    ```

6. **Configuration (`buf.gen.yaml`):** Create `buf.gen.yaml`

    ```yaml
    version: v1
    plugins:
      # Go code generation for Connect
      - plugin: connect-go
        out: internal/infrastructure/rpc/gen
        opt:
          - paths=source_relative
      # OpenAPI v2 generation
      - plugin: openapiv2
        out: third_party/openapi
        opt:
          - allow_merge=true
          - merge_file_name=healthapp/v1/api # Output: third_party/openapi/healthapp/v1/api.swagger.json
          - logtostderr=true # Enable logging for debugging generation
          # - openapi_naming_strategy=fqn # Use fully qualified names
    ```

7. **Configuration (`sqlc.yaml`):** Create `sqlc.yaml`

    ```yaml
    version: "2"
    sql:
      - engine: "postgresql"
        queries: "db/queries" # Directory containing *.sql query files
        schema: "db/migrations" # Directory containing *.sql schema files
        gen:
          go:
            package: "db" # Go package name for generated code
            out: "internal/infrastructure/persistence/postgres/db" # Output directory
            emit_json_tags: true # Add json struct tags
            emit_prepared_queries: false # Use unprepared queries by default (good for pool)
            emit_interface: true # Generate a Querier interface
            emit_exact_table_names: false
            emit_empty_slices: true
            overrides: # Example: Map Postgres UUID to Go's google/uuid
              - db_type: "uuid"
                go_type:
                  import: "github.com/google/uuid"
                  type: "UUID"
              - db_type: "timestamptz"
                go_type:
                  import: "time"
                  type: "Time"
              - db_type: "numeric" # Example for weight/fat percentage
                go_type:
                  # Consider using a dedicated decimal type if precision is critical
                  # or simply float64 if acceptable
                  type: "float64"
    ```

8. **`.gitignore`:** Create `.gitignore` with standard Go entries, plus:

    ```
    # Config
    configs/.env
    configs/config.local.yaml

    # Generated code (re-generate in CI/build)
    internal/infrastructure/persistence/postgres/db/
    internal/infrastructure/rpc/gen/
    third_party/openapi/

    # Build artifacts
    /healthapp-server

    # OS generated files
    .DS_Store
    *.swp
    ```

9. **`docker-compose.yml`:** Create `docker-compose.yml` for local PostgreSQL.

    ```yaml
    version: '3.8'
    services:
      postgres:
        image: postgres:15-alpine
        container_name: healthapp_postgres_dev
        ports:
          - "5432:5432" # Expose standard port
        volumes:
          - postgres_data:/var/lib/postgresql/data
        environment:
          POSTGRES_DB: healthapp_db
          POSTGRES_USER: healthapp_user
          POSTGRES_PASSWORD: verysecretpassword # Use .env or secrets management in real scenario
        healthcheck:
          test: ["CMD-SHELL", "pg_isready -U healthapp_user -d healthapp_db"]
          interval: 5s
          timeout: 5s
          retries: 5

    volumes:
      postgres_data: # Persist data locally between runs
    ```

10. **`Makefile`:** Create a `Makefile` for common commands.

    ```makefile
    .PHONY: help proto sqlc migrate-up migrate-down db-start db-stop test build run clean

    DB_URL=postgres://healthapp_user:verysecretpassword@localhost:5432/healthapp_db?sslmode=disable
    MIGRATIONS_DIR=./db/migrations

    help: ## Display this help screen
            @grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

    proto: ## Generate connect-rpc Go code and OpenAPI spec from proto files
            @echo "Generating Protobuf/Connect code and OpenAPI spec..."
            buf generate api/proto

    sqlc: ## Generate Go code from SQL queries using sqlc
            @echo "Generating sqlc code..."
            sqlc generate

    migrate-create: NAME?=new_migration ## Create new migration files (e.g., make migrate-create NAME=add_indexes)
            @echo "Creating migration: $(NAME)"
            migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)

    migrate-up: ## Apply all database migrations
            @echo "Applying migrations..."
            migrate -database "$(DB_URL)" -path $(MIGRATIONS_DIR) up

    migrate-down: ## Revert the last database migration
            @echo "Reverting last migration..."
            migrate -database "$(DB_URL)" -path $(MIGRATIONS_DIR) down 1

    db-start: ## Start PostgreSQL container via Docker Compose
            @echo "Starting PostgreSQL container..."
            docker compose up -d postgres

    db-stop: ## Stop PostgreSQL container
            @echo "Stopping PostgreSQL container..."
            docker compose down

    test: ## Run Go tests
            @echo "Running tests..."
            go test -v -race ./...

    build: ## Build the server binary
            @echo "Building server..."
            go build -o healthapp-server ./cmd/server

    run: build ## Build and run the server
            @echo "Running server..."
            ./healthapp-server

    clean: ## Clean generated files and build artifacts
            @echo "Cleaning generated files and build artifacts..."
            rm -f healthapp-server
            rm -rf internal/infrastructure/persistence/postgres/db/*
            rm -rf internal/infrastructure/rpc/gen/*
            rm -rf third_party/openapi/*
            # Keep .gitkeep files if any
            find internal/infrastructure/persistence/postgres/db/ -type f ! -name '.gitkeep' -delete
            find internal/infrastructure/rpc/gen/ -type f ! -name '.gitkeep' -delete
            find third_party/openapi/ -type f ! -name '.gitkeep' -delete


    # Combine setup steps
    setup-tools:
            go install github.com/bufbuild/buf/cmd/buf@latest
            go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
            go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
            go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
            go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

    generate-all: proto sqlc ## Generate all code (protobuf, connect, sqlc)

    init-db: db-start ## Initialize DB: Start container & run migrations
            @echo "Waiting for DB to be ready..."
            @until docker compose exec postgres pg_isready -U healthapp_user -d healthapp_db -q; do \
                    sleep 1; \
            done
            $(MAKE) migrate-up

    ```

    *(Usage: `make proto`, `make sqlc`, `make migrate-up`, `make run`, `make init-db`)*

11. **GitHub Actions & Release Please:**
    * Create `.github/workflows/ci.yml` for running linters, tests, code generation checks on PRs/pushes.
    * Create `.github/workflows/release-please.yml` based on Google's `release-please-action` documentation for automated Go releases tagged based on conventional commits. Configure it to trigger on pushes to the `main` branch.

---

**2. Database Design (PostgreSQL & sqlc)**

1. **Schema Migrations (`db/migrations/*.sql`):**
    * Create the first migration: `make migrate-create NAME=create_initial_tables`
    * Edit `db/migrations/000001_create_initial_tables.up.sql`:

        ```sql
        -- Enable UUID generation if not already enabled
        CREATE EXTENSION IF NOT EXISTS "pgcrypto";

        -- Users table (linked via Auth0 subject claim)
        CREATE TABLE users (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            auth0_sub TEXT UNIQUE NOT NULL, -- Auth0 subject claim (e.g., "auth0|...")
            created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
        );
        CREATE INDEX idx_users_auth0_sub ON users (auth0_sub);

        -- Body composition records
        CREATE TABLE body_records (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            user_id UUID NOT NULL,
            date DATE NOT NULL, -- Date of the record
            weight_kg NUMERIC(5, 2), -- Weight in kilograms, e.g., 75.50
            body_fat_percentage NUMERIC(4, 2), -- Body fat percentage, e.g., 15.25
            created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
            CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
            UNIQUE (user_id, date) -- Allow only one record per user per day
        );
        CREATE INDEX idx_body_records_user_date ON body_records (user_id, date DESC);

        -- Exercise records
        CREATE TABLE exercise_records (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            user_id UUID NOT NULL,
            exercise_name TEXT NOT NULL, -- e.g., "Running", "Weight Lifting"
            duration_minutes INTEGER, -- Duration in minutes
            calories_burned INTEGER, -- Estimated calories burned
            recorded_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, -- When the exercise was performed/logged
            created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
            CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
        );
        CREATE INDEX idx_exercise_records_user_recorded_at ON exercise_records (user_id, recorded_at DESC);

        -- Diary entries
        CREATE TABLE diary_entries (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            user_id UUID NOT NULL,
            title TEXT, -- Optional title
            content TEXT NOT NULL, -- The main diary text
            entry_date DATE NOT NULL, -- Date the diary entry pertains to
            created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
            CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
        );
        CREATE INDEX idx_diary_entries_user_entry_date ON diary_entries (user_id, entry_date DESC);

        -- Columns/Articles
        CREATE TABLE columns (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            title TEXT NOT NULL,
            content TEXT NOT NULL,
            category TEXT, -- e.g., "Diet", "Exercise", "Mental Health"
            tags TEXT[], -- Array of tags
            published_at TIMESTAMPTZ, -- Nullable, only show if not null and in the past
            created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
        );
        CREATE INDEX idx_columns_published_at ON columns (published_at DESC NULLS LAST);
        CREATE INDEX idx_columns_category ON columns (category);
        CREATE INDEX idx_columns_tags ON columns USING GIN (tags); -- GIN index for array searching

        -- Function to automatically update 'updated_at' timestamps
        CREATE OR REPLACE FUNCTION trigger_set_timestamp()
        RETURNS TRIGGER AS $$
        BEGIN
          NEW.updated_at = NOW();
          RETURN NEW;
        END;
        $$ LANGUAGE plpgsql;

        -- Apply the trigger to relevant tables
        CREATE TRIGGER set_timestamp_users BEFORE UPDATE ON users FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();
        CREATE TRIGGER set_timestamp_body_records BEFORE UPDATE ON body_records FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();
        CREATE TRIGGER set_timestamp_exercise_records BEFORE UPDATE ON exercise_records FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();
        CREATE TRIGGER set_timestamp_diary_entries BEFORE UPDATE ON diary_entries FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();
        CREATE TRIGGER set_timestamp_columns BEFORE UPDATE ON columns FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();

        ```

    * Create the corresponding `db/migrations/000001_create_initial_tables.down.sql`:

        ```sql
        DROP TRIGGER IF EXISTS set_timestamp_columns ON columns;
        DROP TRIGGER IF EXISTS set_timestamp_diary_entries ON diary_entries;
        DROP TRIGGER IF EXISTS set_timestamp_exercise_records ON exercise_records;
        DROP TRIGGER IF EXISTS set_timestamp_body_records ON body_records;
        DROP TRIGGER IF EXISTS set_timestamp_users ON users;

        DROP FUNCTION IF EXISTS trigger_set_timestamp();

        DROP TABLE IF EXISTS columns;
        DROP TABLE IF EXISTS diary_entries;
        DROP TABLE IF EXISTS exercise_records;
        DROP TABLE IF EXISTS body_records;
        DROP TABLE IF EXISTS users;

        DROP EXTENSION IF EXISTS "pgcrypto";
        ```

    * Initialize the local DB: `make init-db` (This runs `db-start` and `migrate-up`).

2. **SQL Queries (`db/queries/*.sql`):** Create files like `user.sql`, `body_record.sql`, etc.
    * `db/queries/user.sql`:

        ```sql
        -- name: CreateUser :one
        INSERT INTO users (auth0_sub)
        VALUES ($1)
        RETURNING *;

        -- name: GetUserByID :one
        SELECT * FROM users
        WHERE id = $1 LIMIT 1;

        -- name: GetUserByAuth0Sub :one
        SELECT * FROM users
        WHERE auth0_sub = $1 LIMIT 1;
        ```

    * `db/queries/body_record.sql`:

        ```sql
        -- name: CreateBodyRecord :one
        INSERT INTO body_records (user_id, date, weight_kg, body_fat_percentage)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (user_id, date) DO UPDATE SET
            weight_kg = EXCLUDED.weight_kg,
            body_fat_percentage = EXCLUDED.body_fat_percentage,
            updated_at = CURRENT_TIMESTAMP
        RETURNING *;

        -- name: ListBodyRecordsByUser :many
        SELECT * FROM body_records
        WHERE user_id = $1
        ORDER BY date DESC
        LIMIT $2 OFFSET $3; -- For pagination

        -- name: ListBodyRecordsByUserDateRange :many
        SELECT * FROM body_records
        WHERE user_id = $1 AND date >= $2 AND date <= $3
        ORDER BY date ASC;
        ```

    * *(Create similar `.sql` files and queries for `exercise_records`, `diary_entries`, `columns` covering necessary CRUD and listing operations)*

3. **Generate SQLC Code:** Run `make sqlc`. Verify code generation in `internal/infrastructure/persistence/postgres/db/`.

---

**3. API Design (connect-rpc & Proto)**

1. **Proto Definitions (`api/proto/healthapp/v1/*.proto`):**
    * `api/proto/healthapp/v1/common.proto`:

        ```protobuf
        syntax = "proto3";

        package healthapp.v1;

        option go_package = "github.com/<your-org>/healthapp-backend/internal/infrastructure/rpc/gen/healthapp/v1;healthappv1";

        // Standard pagination request
        message PageRequest {
          int32 page_size = 1; // Number of items per page (0 for default)
          int32 page_number = 2; // Page number (1-based)
        }

        // Standard pagination response
        message PageResponse {
          int32 total_items = 1;
          int32 total_pages = 2;
          int32 current_page = 3;
        }
        ```

    * `api/proto/healthapp/v1/user.proto`:

        ```protobuf
        syntax = "proto3";

        package healthapp.v1;

        import "google/protobuf/timestamp.proto";
        import "google/api/annotations.proto"; // For potential future REST mapping

        option go_package = "github.com/<your-org>/healthapp-backend/internal/infrastructure/rpc/gen/healthapp/v1;healthappv1";

        // User resource representation
        message User {
          string id = 1; // UUID string
          google.protobuf.Timestamp created_at = 2;
          google.protobuf.Timestamp updated_at = 3;
          // Avoid exposing auth0_sub directly in APIs if possible
        }

        // Service for user-related operations (currently minimal)
        service UserService {
          // GetAuthenticatedUser retrieves the profile for the currently logged-in user.
          // Requires authentication.
          rpc GetAuthenticatedUser(GetAuthenticatedUserRequest) returns (GetAuthenticatedUserResponse) {
              // Example of adding potential future REST mapping (optional)
              option (google.api.http) = {
                  get: "/v1/users/me"
              };
          }
        }

        message GetAuthenticatedUserRequest {}

        message GetAuthenticatedUserResponse {
          User user = 1;
        }
        ```

    * `api/proto/healthapp/v1/body_record.proto`:

        ```protobuf
        syntax = "proto3";

        package healthapp.v1;

        import "google/protobuf/timestamp.proto";
        import "google/protobuf/wrappers.proto"; // For optional fields
        import "healthapp/v1/common.proto";

        option go_package = "github.com/<your-org>/healthapp-backend/internal/infrastructure/rpc/gen/healthapp/v1;healthappv1";

        message BodyRecord {
          string id = 1; // UUID string
          string user_id = 2; // UUID string
          string date = 3; // Date string in "YYYY-MM-DD" format
          google.protobuf.DoubleValue weight_kg = 4; // Optional: Use wrappers for nullability
          google.protobuf.DoubleValue body_fat_percentage = 5; // Optional
          google.protobuf.Timestamp created_at = 6;
          google.protobuf.Timestamp updated_at = 7;
        }

        service BodyRecordService {
          // Create or update a body record for a specific date.
          // Requires authentication.
          rpc CreateBodyRecord(CreateBodyRecordRequest) returns (CreateBodyRecordResponse);

          // List body records for the authenticated user, paginated.
          // Requires authentication.
          rpc ListBodyRecords(ListBodyRecordsRequest) returns (ListBodyRecordsResponse);

          // List body records for a specific date range.
          // Requires authentication.
          rpc GetBodyRecordsByDateRange(GetBodyRecordsByDateRangeRequest) returns (GetBodyRecordsByDateRangeResponse);
        }

        message CreateBodyRecordRequest {
          string date = 1; // "YYYY-MM-DD"
          google.protobuf.DoubleValue weight_kg = 2;
          google.protobuf.DoubleValue body_fat_percentage = 3;
        }

        message CreateBodyRecordResponse {
          BodyRecord body_record = 1;
        }

        message ListBodyRecordsRequest {
          PageRequest pagination = 1;
        }

        message ListBodyRecordsResponse {
          repeated BodyRecord body_records = 1;
          PageResponse pagination = 2;
        }

        message GetBodyRecordsByDateRangeRequest {
            string start_date = 1; // "YYYY-MM-DD" inclusive
            string end_date = 2; // "YYYY-MM-DD" inclusive
        }

        message GetBodyRecordsByDateRangeResponse {
            repeated BodyRecord body_records = 1;
        }
        ```

    * *(Create similar `.proto` files for `ExerciseService`, `DiaryService`, `ColumnService` defining messages and RPC methods according to application features)*

2. **Lint & Check Breaking Changes:**
    * `buf lint api/proto`
    * *(In CI)* `buf breaking --against '.git#tag=vX.Y.Z'` (Compare against last release tag)

3. **Generate Connect & OpenAPI Code:** Run `make proto`. Verify code in `internal/infrastructure/rpc/gen/` and `third_party/openapi/`.

---

**4. Domain Modeling (DDD)**

1. **Entities (`internal/domain/*.go`):** Define plain Go structs.
    * `internal/domain/user.go`:

        ```go
        package domain

        import (
            "time"
            "github.com/google/uuid"
        )

        type User struct {
            ID        uuid.UUID
            Auth0Sub  string // Keep internal, don't expose in API directly
            CreatedAt time.Time
            UpdatedAt time.Time
        }
        ```

    * `internal/domain/body_record.go`:

        ```go
        package domain

        import (
            "time"
            "github.com/google/uuid"
        )

        // Using pointers for optional numeric fields to distinguish between 0 and null/not set
        type BodyRecord struct {
            ID                uuid.UUID
            UserID            uuid.UUID
            Date              time.Time // Store as time.Time (YYYY-MM-DD 00:00:00 UTC)
            WeightKg          *float64
            BodyFatPercentage *float64
            CreatedAt         time.Time
            UpdatedAt         time.Time
        }

        // Example domain validation
        func (br *BodyRecord) Validate() error {
            // Add validation rules here, e.g., range checks for weight/percentage
            return nil
        }
        ```

    * *(Define similar structs for `ExerciseRecord`, `DiaryEntry`, `Column`)*

2. **Repository Interfaces (`internal/domain/*_repository.go`):** Define contracts.
    * `internal/domain/user_repository.go`:

        ```go
        package domain

        import (
            "context"
            "github.com/google/uuid"
        )

        // ErrUserNotFound is returned when a user is not found.
        var ErrUserNotFound = errors.New("user not found") // Example error

        type UserRepository interface {
            // Create creates a new user record.
            Create(ctx context.Context, user *User) error
            // FindByID retrieves a user by their internal UUID. Returns ErrUserNotFound if not found.
            FindByID(ctx context.Context, id uuid.UUID) (*User, error)
            // FindByAuth0Sub retrieves a user by their Auth0 Subject claim. Returns ErrUserNotFound if not found.
            FindByAuth0Sub(ctx context.Context, auth0Sub string) (*User, error)
        }
        ```

    * `internal/domain/body_record_repository.go`:

        ```go
        package domain

        import (
            "context"
            "time"
            "github.com/google/uuid"
        )

        type BodyRecordRepository interface {
            // Save creates a new body record or updates an existing one based on UserID and Date.
            Save(ctx context.Context, record *BodyRecord) (*BodyRecord, error)
            // FindByUser retrieves paginated body records for a user.
            FindByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*BodyRecord, error)
            // FindByUserAndDateRange retrieves body records for a user within a specific date range.
            FindByUserAndDateRange(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]*BodyRecord, error)
            // CountByUser returns the total number of body records for a user.
            CountByUser(ctx context.Context, userID uuid.UUID) (int64, error)
        }
        ```

    * *(Define similar repository interfaces for `ExerciseRecord`, `DiaryEntry`, `Column`)*

---

**5. Implementation**

1. **Configuration (`internal/infrastructure/config/config.go`):**
    * Define a `Config` struct mirroring `configs/config.yaml` structure.
    * Use `viper` to load from file (`configs/config.yaml`, `configs/config.local.yaml`) and environment variables (prefix `HEALTHAPP_`).
    * Include fields for `Server.Port`, `Database.URL`, `Auth0.Domain`, `Auth0.Audience`, `Auth0.JWKSURI`.
    * Provide a `LoadConfig()` function.

2. **Logging (`internal/infrastructure/log/logger.go`):**
    * Initialize `slog` (e.g., `slog.New(slog.NewJSONHandler(os.Stdout, nil))`).
    * Provide a function to get the configured logger instance.

3. **Persistence Implementation (`internal/infrastructure/persistence/postgres/`):**
    * `postgres.go`: Contains function `NewDBPool(config *config.DatabaseConfig)` to create and return a `*pgxpool.Pool`.
    * `user_repo_pg.go`:

        ```go
        package postgres

        import (
            // ... imports including domain, sqlc gen pkg, context, pgxpool, errors, uuid
            db "github.com/<your-org>/healthapp-backend/internal/infrastructure/persistence/postgres/db" // Alias sqlc gen
            "github.com/<your-org>/healthapp-backend/internal/domain"
            "github.com/jackc/pgx/v5"
            "github.com/jackc/pgx/v5/pgconn"
        )

        type pgUserRepository struct {
            db *pgxpool.Pool
            q  *db.Queries // sqlc generated queries
        }

        func NewPgUserRepository(pool *pgxpool.Pool) domain.UserRepository {
            return &pgUserRepository{db: pool, q: db.New(pool)}
        }

        // Implement domain.UserRepository interface methods using q.Method(...)
        func (r *pgUserRepository) Create(ctx context.Context, user *domain.User) error {
            _, err := r.q.CreateUser(ctx, user.Auth0Sub)
            if err != nil {
                 // Check for unique constraint violation, etc.
                var pgErr *pgconn.PgError
                if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
                    return errors.New("user with this auth0 sub already exists") // Wrap error appropriately
                }
                return fmt.Errorf("failed to create user: %w", err)
            }
            // sqlc CreateUser returns the created user, maybe update the input user object ID?
            // Or Fetch after create if needed by domain logic (depends on sqlc query)
            return nil
        }

        func (r *pgUserRepository) FindByAuth0Sub(ctx context.Context, auth0Sub string) (*domain.User, error) {
            dbUser, err := r.q.GetUserByAuth0Sub(ctx, auth0Sub)
            if err != nil {
                if errors.Is(err, pgx.ErrNoRows) {
                    return nil, domain.ErrUserNotFound // Use defined domain error
                }
                return nil, fmt.Errorf("failed to get user by auth0 sub: %w", err)
            }
            // Convert sqlc db.User to domain.User
            return &domain.User{
                ID:        dbUser.ID,
                Auth0Sub:  dbUser.Auth0Sub,
                CreatedAt: dbUser.CreatedAt,
                UpdatedAt: dbUser.UpdatedAt,
            }, nil
        }
         // Implement FindByID similarly...

        // --- Helper for potential conversion ---
        // func toDomainUser(dbUser db.User) *domain.User { ... }
        ```

    * *(Implement other repository interfaces (`body_record_repo_pg.go`, etc.) similarly, handling type conversions between `sqlc` generated structs and `domain` structs, and mapping `pgx.ErrNoRows` to domain-specific `ErrNotFound` errors)*

4. **Auth Interceptor (`internal/infrastructure/auth/auth_interceptor.go`):**
    * Define a type `authInterceptor`.
    * Function `NewAuthInterceptor(cfg *config.Auth0Config)`:
        * Creates JWKS cache using `keyfunc.NewDefault([]string{cfg.JWKSURI})`.
        * Returns an interceptor function `func(next connect.UnaryFunc) connect.UnaryFunc`.
    * The interceptor function:
        * Checks if the RPC method requires auth (e.g., based on naming convention or annotation - simple approach: protect all by default unless explicitly marked public).
        * Extracts "Bearer <token>" from `req.Header().Get("Authorization")`. Returns `connect.CodeUnauthenticated` if missing/malformed.
        * Parses the token using `jwt.Parse` with the JWKS `keyfunc`.
        * Validates claims: `iss` (issuer - `cfg.Domain`), `aud` (audience - `cfg.Audience`). Returns `connect.CodeUnauthenticated` on validation failure.
        * Extracts the `sub` claim (Auth0 User ID).
        * **User Synchronization:**
            * Inject `domain.UserRepository` into the interceptor setup.
            * Call `userRepo.FindByAuth0Sub(ctx, sub)`.
            * If `domain.ErrUserNotFound`, call `userRepo.Create(ctx, &domain.User{Auth0Sub: sub})`. Handle creation error.
            * If found or successfully created, get the internal `user.ID` (UUID).
            * Add the internal User ID to the context: `ctx = context.WithValue(ctx, userContextKey, user.ID)`. Define `userContextKey` as a private context key type.
        * Calls `next(ctx, req)`.
        * Handles errors from `next` or validation.

5. **Application Services (`internal/application/*.go`):**
    * `body_record_service.go`:

        ```go
        package application

        import (
            // ... imports including domain, context, uuid, time, log/slog
            "github.com/<your-org>/healthapp-backend/internal/domain"
        )

        type BodyRecordService interface {
             CreateOrUpdateBodyRecord(ctx context.Context, userID uuid.UUID, date time.Time, weight *float64, fatPercent *float64) (*domain.BodyRecord, error)
             GetBodyRecordsForUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*domain.BodyRecord, int64, error) // Returns records and total count
             GetBodyRecordsForUserDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]*domain.BodyRecord, error)
        }

        type bodyRecordServiceImpl struct {
             repo domain.BodyRecordRepository
             log  *slog.Logger
        }

        func NewBodyRecordService(repo domain.BodyRecordRepository, log *slog.Logger) BodyRecordService {
            return &bodyRecordServiceImpl{repo: repo, log: log}
        }

        func (s *bodyRecordServiceImpl) CreateOrUpdateBodyRecord(ctx context.Context, userID uuid.UUID, date time.Time, weight *float64, fatPercent *float64) (*domain.BodyRecord, error) {
            record := &domain.BodyRecord{
                UserID:            userID,
                Date:              date, // Ensure date is UTC noon or start of day
                WeightKg:          weight,
                BodyFatPercentage: fatPercent,
                // ID, CreatedAt, UpdatedAt will be set by DB/repo
            }

            // Optional: Call domain validation
            if err := record.Validate(); err != nil {
                 s.log.WarnContext(ctx, "Validation failed for body record", "userID", userID, "error", err)
                 return nil, fmt.Errorf("invalid body record data: %w", err) // Return validation error
            }

            s.log.InfoContext(ctx, "Saving body record", "userID", userID, "date", date)
            savedRecord, err := s.repo.Save(ctx, record)
            if err != nil {
                 s.log.ErrorContext(ctx, "Failed to save body record", "userID", userID, "error", err)
                 return nil, fmt.Errorf("could not save body record: %w", err) // Wrap repo error
            }
            return savedRecord, nil
        }

        func (s *bodyRecordServiceImpl) GetBodyRecordsForUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*domain.BodyRecord, int64, error) {
            if pageSize <= 0 || pageSize > 100 { // Add default/max page size logic
                pageSize = 20
            }
            if page <= 0 {
                page = 1
            }
            offset := (page - 1) * pageSize

            s.log.InfoContext(ctx, "Fetching body records for user", "userID", userID, "page", page, "pageSize", pageSize)
            records, err := s.repo.FindByUser(ctx, userID, pageSize, offset)
            if err != nil {
                 s.log.ErrorContext(ctx, "Failed to fetch body records", "userID", userID, "error", err)
                 return nil, 0, fmt.Errorf("could not fetch body records: %w", err)
            }

            total, err := s.repo.CountByUser(ctx, userID)
             if err != nil {
                 s.log.ErrorContext(ctx, "Failed to count body records", "userID", userID, "error", err)
                 return nil, 0, fmt.Errorf("could not count body records: %w", err)
             }

            return records, total, nil
        }
         // Implement GetBodyRecordsForUserDateRange similarly...
        ```

    * *(Implement other application services (`UserService`, `ExerciseService`, etc.) injecting necessary repositories and logger)*

6. **RPC Handlers (`internal/infrastructure/rpc/handlers/*.go`):**
    * `user_handler.go`:

        ```go
        package handlers

        import (
            // ... imports including connect, context, generated protos, application services, uuid, slog
            v1 "github.com/<your-org>/healthapp-backend/internal/infrastructure/rpc/gen/healthapp/v1"
            "github.com/<your-org>/healthapp-backend/internal/application"
            "github.com/<your-org>/healthapp-backend/internal/infrastructure/auth" // For context key
            "google.golang.org/protobuf/types/known/timestamppb"

        )

        type UserHandler struct {
             userApp application.UserService // Inject application service
             log     *slog.Logger
        }

        func NewUserHandler(userApp application.UserService, log *slog.Logger) *UserHandler {
             return &UserHandler{userApp: userApp, log: log}
        }

        func (h *UserHandler) GetAuthenticatedUser(ctx context.Context, req *connect.Request[v1.GetAuthenticatedUserRequest]) (*connect.Response[v1.GetAuthenticatedUserResponse], error) {
            userID, ok := ctx.Value(auth.UserContextKey).(uuid.UUID)
            if !ok {
                 h.log.ErrorContext(ctx, "User ID not found in context")
                 return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("user not authenticated"))
            }

            h.log.InfoContext(ctx, "Handling GetAuthenticatedUser request", "userID", userID)
            user, err := h.userApp.GetUserProfile(ctx, userID) // Assuming UserService has GetUserProfile
            if err != nil {
                h.log.ErrorContext(ctx, "Failed to get user profile", "userID", userID, "error", err)
                 // Map domain/app errors to connect errors
                 if errors.Is(err, domain.ErrUserNotFound) { // Example mapping
                    return nil, connect.NewError(connect.CodeNotFound, errors.New("user profile not found"))
                 }
                 return nil, connect.NewError(connect.CodeInternal, errors.New("failed to retrieve user profile"))
            }

            // Convert domain.User to v1.User
            protoUser := &v1.User{
                 Id:        user.ID.String(),
                 CreatedAt: timestamppb.New(user.CreatedAt),
                 UpdatedAt: timestamppb.New(user.UpdatedAt),
            }

            res := connect.NewResponse(&v1.GetAuthenticatedUserResponse{
                 User: protoUser,
            })
            return res, nil
        }
        ```

    * `body_record_handler.go`:
        * Implement `v1connect.BodyRecordServiceHandler` interface.
        * Inject `application.BodyRecordService`.
        * In methods like `CreateBodyRecord`:
            * Extract `userID` from context.
            * Parse date string (`req.Msg.Date`) into `time.Time`. Handle parsing errors -> `connect.CodeInvalidArgument`.
            * Convert protobuf `DoubleValue` to `*float64`.
            * Call the application service `h.bodyApp.CreateOrUpdateBodyRecord(...)`.
            * Handle application service errors, mapping them to `connect.Code...`.
            * Convert the returned `domain.BodyRecord` back to `v1.BodyRecord` protobuf message.
            * Return `connect.NewResponse(...)`.
        * In methods like `ListBodyRecords`:
            * Extract `userID` from context.
            * Get pagination params from `req.Msg.Pagination`. Apply defaults/limits.
            * Call `h.bodyApp.GetBodyRecordsForUser(...)`.
            * Convert returned `[]*domain.BodyRecord` to `[]*v1.BodyRecord`.
            * Construct `v1.PageResponse` from total count and request params.
            * Return `connect.NewResponse(...)`.
    * *(Implement other handlers (`ExerciseHandler`, `DiaryHandler`, `ColumnHandler`) similarly, performing request validation, context extraction, application service calls, error mapping, and domain-to-proto conversion)*

7. **Server Entrypoint (`cmd/server/main.go`):**

    ```go
    package main

    import (
        "context"
        "errors"
        "fmt"
        "log/slog"
        "net/http"
        "os"
        "os/signal"
        "syscall"
        "time"

        "connectrpc.com/connect"
        "connectrpc.com/otelconnect" // Optional: for OpenTelemetry tracing
        "golang.org/x/net/http2"
        "golang.org/x/net/http2/h2c"

        // App imports
        "github.com/<your-org>/healthapp-backend/internal/application"
        "github.com/<your-org>/healthapp-backend/internal/infrastructure/auth"
        "github.com/<your-org>/healthapp-backend/internal/infrastructure/config"
        applog "github.com/<your-org>/healthapp-backend/internal/infrastructure/log" // aliased
        "github.com/<your-org>/healthapp-backend/internal/infrastructure/persistence/postgres"
        rpcgen "github.com/<your-org>/healthapp-backend/internal/infrastructure/rpc/gen/healthapp/v1" // aliased
        "github.com/<your-org>/healthapp-backend/internal/infrastructure/rpc/gen/healthapp/v1/healthappv1connect" // Connect generated paths
        "github.com/<your-org>/healthapp-backend/internal/infrastructure/rpc/handlers"

        // DB Driver
         _ "github.com/jackc/pgx/v5/stdlib" // Register pgx driver for migrate (optional if using migrate CLI directly)

    )

    func main() {
        logger := applog.NewLogger() // Initialize structured logger
        logger.Info("Starting server...")

        cfg, err := config.LoadConfig("./configs") // Load config from ./configs dir
        if err != nil {
            logger.Error("Failed to load configuration", "error", err)
            os.Exit(1)
        }

        // --- Dependency Injection Setup ---
        // Database Pool
        dbPool, err := postgres.NewDBPool(&cfg.Database)
        if err != nil {
            logger.Error("Failed to connect to database", "error", err)
            os.Exit(1)
        }
        defer dbPool.Close()
        logger.Info("Database connection pool established")

        // Repositories
        userRepo := postgres.NewPgUserRepository(dbPool)
        bodyRecordRepo := postgres.NewPgBodyRecordRepository(dbPool)
        // ... initialize other repositories

        // Application Services
        // userService := application.NewUserService(userRepo, logger) // Uncomment when UserService exists
        bodyRecordService := application.NewBodyRecordService(bodyRecordRepo, logger)
        // ... initialize other services

        // Middlewares/Interceptors
        // Optional: Add recovery interceptor, logging interceptor, metrics, tracing
        authInterceptor, err := auth.NewAuthInterceptor(&cfg.Auth0, userRepo, logger) // Pass repo for user sync
         if err != nil {
             logger.Error("Failed to create auth interceptor", "error", err)
             os.Exit(1)
         }
        interceptors := connect.WithInterceptors(
             // otelconnect.NewInterceptor(), // Optional: OpenTelemetry
             authInterceptor,
             // Add more interceptors here (logging, metrics, recovery)
        )


        // RPC Handlers
        // userHandler := handlers.NewUserHandler(userService, logger) // Uncomment when UserService exists
        bodyRecordHandler := handlers.NewBodyRecordHandler(bodyRecordService, logger)
        // ... initialize other handlers

        // Router & Server
        mux := http.NewServeMux()
        basePath := fmt.Sprintf("/api/%s/", rpcgen.HealthappV1Version) // Base path e.g., /api/healthapp.v1/

        // Register Connect handlers
        // mux.Handle(healthappv1connect.NewUserServiceHandler(userHandler, interceptors)) // Uncomment when handler exists
        mux.Handle(healthappv1connect.NewBodyRecordServiceHandler(bodyRecordHandler, interceptors))
        // ... register other handlers

        // Optional: Serve OpenAPI spec & Swagger UI
        mux.Handle("/openapi/", http.StripPrefix("/openapi/", http.FileServer(http.Dir("./third_party/openapi"))))
        // Consider embedding swagger UI or using a dedicated handler
        // mux.Handle("/swagger/", http.StripPrefix("/swagger/", http.FileServer(http.Dir("./third_party/swagger-ui/dist")))) // If files are present

        addr := fmt.Sprintf(":%s", cfg.Server.Port)
        logger.Info("Server listening", "address", addr)

        // Use h2c for local development (HTTP/1.1 -> HTTP/2 upgrade) without TLS
        // For production, use http.ListenAndServeTLS with http2.ConfigureServer
        server := &http.Server{
            Addr:         addr,
            Handler:      h2c.NewHandler(mux, &http2.Server{}),
            ReadTimeout:  5 * time.Second,
            WriteTimeout: 10 * time.Second,
            IdleTimeout:  120 * time.Second,
        }

        // Graceful Shutdown Handling
        stopChan := make(chan os.Signal, 1)
        signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

        go func() {
            if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
                logger.Error("Server failed to start", "error", err)
                os.Exit(1)
            }
        }()

        // Wait for interrupt signal
        <-stopChan
        logger.Info("Shutdown signal received, initiating graceful shutdown...")

        // Shutdown context with timeout
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
        defer cancel()

        if err := server.Shutdown(shutdownCtx); err != nil {
            logger.Error("Server graceful shutdown failed", "error", err)
            os.Exit(1)
        }

        logger.Info("Server shutdown gracefully")
    }
    ```

---

**6. OpenAPI Documentation**

1. **Generation:** Already configured in `buf.gen.yaml` and `make proto`. The output is `third_party/openapi/healthapp/v1/api.swagger.json`.
2. **Serving:** The `cmd/server/main.go` includes a basic example handler to serve the generated JSON file at `/openapi/healthapp/v1/api.swagger.json`.
3. **(Optional) Viewing:** Download Swagger UI static files (`https://github.com/swagger-api/swagger-ui/releases`) into `third_party/swagger-ui/dist/`. Add a handler in `main.go` (like the commented-out example) to serve these files at `/swagger/`. Access `http://localhost:<port>/swagger/?url=/openapi/healthapp/v1/api.swagger.json` in a browser.

---

**7. Testing**

1. **Unit Tests (`*_test.go`):**
    * Test domain entity validation logic (`internal/domain/body_record_test.go`).
    * Test application service logic (`internal/application/body_record_service_test.go`) using mocked repositories (`testify/mock`). Mock repository methods to return predefined data or errors and assert service behavior.
2. **Integration Tests (`*_integration_test.go`):**
    * Use `testcontainers-go` to spin up a PostgreSQL container specifically for tests.
    * Write tests that interact with the *real* repository implementations (`internal/infrastructure/persistence/postgres/*_repo_pg_test.go`). Seed the test database, call repository methods, and assert results against the DB state.
    * Test the Auth interceptor logic with mock JWTs or against a test Auth0 tenant if feasible.
    * Test the RPC handlers by setting up a test server (`httptest.NewServer`) with real services/mocked services and using a generated Connect client to make calls.

---

**8. Key Considerations & Decisions (Summary)**

* **Error Handling:** Use custom error variables (e.g., `domain.ErrNotFound`) and `fmt.Errorf("...: %w", err)` for wrapping. Map errors consistently to `connect.Code` in the RPC handler layer. Log errors with context.
* **Configuration:** Viper loads from `config.yaml` + `config.local.yaml` + ENV VARS (`HEALTHAPP_...`). Sensitive data (DB pass, Auth0 secrets) should *only* be in `.env` (gitignored) or ENV VARS, never committed.
* **Database Migrations:** Use `golang-migrate/migrate` via `Makefile` commands (`make migrate-create`, `make migrate-up`, `make migrate-down`). Migrations are sequential SQL files.
* **User Synchronization:** Implemented in the Auth Interceptor: On first valid JWT from an unknown `sub`, create a corresponding record in the `users` table. Fetch internal User ID on subsequent requests.
* **Context Propagation:** Pass `context.Context` through all layers (handler -> service -> repository). Use `context.WithValue` *sparingly*, primarily for auth data (user ID).
* **Logging:** Use `log/slog` for structured JSON logging. Include request/correlation IDs if possible (e.g., via middleware). Log key info at `Info` level, errors at `Error` level.
* **CI Pipeline:** GitHub Actions workflow (`ci.yml`) runs `make test`, `buf lint`, `buf breaking`, `make generate-all` (to check for diffs), `go build`. (`release-please.yml`) handles releases.
* **Dependency Injection:** Manual DI shown in `main.go`. Consider `google/wire` for larger applications to automate DI code generation.

```
