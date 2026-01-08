# Integration Tests

This document describes how to run the integration tests for the pf-service.

## Prerequisites

1. **PostgreSQL Database**: You need a PostgreSQL database running for integration tests.
   - Default connection: `host=localhost user=postgres password=test dbname=postgres_test port=5432 sslmode=disable`
   - You can override this by setting the `TEST_POSTGRES_DSN` environment variable

2. **Test Database Setup**: The tests will automatically create and migrate the required tables:
   - `core_app_customuser`
   - `core_app_property`
   - `core_app_property_translation`
   - `core_app_propertyimage`

## Running Tests

### Run all tests:
```bash
go test ./...
```

### Run only integration tests:
```bash
go test -v ./integration_test.go
```

### Run specific test package:
```bash
# Database tests
go test -v ./internal/db/...

# Media download tests
go test -v ./internal/media_download/...

# Integration tests
go test -v -run TestIntegrationSyncFlow
```

### Run with custom database:
```bash
TEST_POSTGRES_DSN="host=localhost user=postgres password=test dbname=test_db port=5432 sslmode=disable" go test -v ./...
```

## Test Structure

### Unit Tests

- **`internal/db/db_test.go`**: Tests for database operations
  - User save/update
  - Property save/update
  - Property image save

- **`internal/media_download/downloader_test.go`**: Tests for image downloading
  - Image download with valid URL
  - Error handling (empty URL, invalid URL, 404, empty response)
  - UUID fallback for missing filenames
  - Media root configuration

### Integration Tests

- **`integration_test.go`**: Full end-to-end sync flow tests
  - Complete sync process from API to database
  - HTTP client mocking
  - Image downloading and storage
  - Database persistence verification
  - Media path configuration validation

## Test Coverage

The integration tests cover:

1. **JWT Token Retrieval**: Mocked token endpoint
2. **User Fetching**: Mocked users API endpoint
3. **Listings Fetching**: Mocked listings API endpoint
4. **User Conversion**: PF user to Django user conversion
5. **Property Creation/Update**: Property save with translations
6. **Image Downloading**: Image download from mock server
7. **Database Persistence**: Verification of saved data
8. **Media Path Configuration**: Validation that paths match Docker setup

## Media Path Configuration

The service is configured to work with the Docker setup:

- **pf-sync container**: `/mhp/media` (mounted from host)
- **Django API container**: `/var/www/app/media` (mounted from same host directory)
- **Default MEDIA_ROOT**: `/mhp/media` (can be overridden with `MEDIA_ROOT` env var)
- **Image storage**: `/mhp/media/property_images/`
- **Returned path**: `property_images/filename.jpg` (relative path for Django)

This ensures that images downloaded by pf-sync are accessible to the Django API through the shared volume mount.

## Notes

- Tests use temporary directories for media files (cleaned up after tests)
- Database tables are truncated before each test run
- HTTP endpoints are mocked using `httptest.Server`
- All tests are designed to be independent and can run in parallel

