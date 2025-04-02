# Test Utilities

This package provides utilities for testing with a real PostgreSQL database in a Docker container.

## Overview

The test utilities in this package allow you to:

1. Spin up a PostgreSQL Docker container for testing
2. Run database migrations to set up the schema
3. Create test data (users, body records, etc.)
4. Clean up after tests

## Usage

### Setting Up a Test Database

```go
// Set up the test database
testDB := testutil.SetupTestDatabase(t)
defer testDB.TeardownTestDatabase(t)
```

This will:
- Start a PostgreSQL container
- Run migrations to set up the schema
- Return a `TestDatabase` struct with connections to the database

### Creating Test Data

```go
// Create a test user
ctx := context.Background()
userID, err := testutil.CreateTestUser(ctx, testDB.DB)
if err != nil {
    t.Fatalf("Failed to create test user: %v", err)
}

// Create a test body record
weight := 75.5
bodyFat := 15.5
record, err := testutil.CreateTestBodyRecord(ctx, testDB.DB, userID, time.Now(), &weight, &bodyFat)
if err != nil {
    t.Fatalf("Failed to create test body record: %v", err)
}
```

### Creating a Repository

```go
// Create a repository for testing
repo := testutil.NewBodyRecordRepository(testDB.Pool)
```

## Integration Tests

Integration tests that use a real database can be conditionally enabled by setting the `RUN_INTEGRATION_TESTS` environment variable:

```go
// Skip integration tests if not explicitly enabled
if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
    t.Skip("Skipping integration test. Set RUN_INTEGRATION_TESTS=true to run")
}
```

To run integration tests:

```bash
RUN_INTEGRATION_TESTS=true go test ./...
```

## Benefits Over Mocks

Using a real database for testing has several advantages:

1. Tests are more realistic and closer to production behavior
2. No need to update mocks when the database schema changes
3. Can test actual SQL queries and database interactions
4. Easier to maintain as the codebase evolves
5. Can catch issues that might not be apparent with mocks

The trade-off is that tests may run slightly slower, but the increased reliability and reduced maintenance make it worthwhile.
