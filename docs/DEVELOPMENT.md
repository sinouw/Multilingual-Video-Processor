# Development Guide

This guide covers local development setup, workflow, and best practices for contributing to the Multilingual Video Processor.

## Local Development Setup

### Prerequisites

1. **Go 1.23+**: Install from [golang.org](https://golang.org/dl/)
2. **FFmpeg**: Required for video processing
   - macOS: `brew install ffmpeg`
   - Linux: `apt-get install ffmpeg` or `yum install ffmpeg`
   - Windows: Download from [ffmpeg.org](https://ffmpeg.org/download.html)
3. **Google Cloud SDK**: For deployment (optional for local dev)
4. **Functions Framework**: For local testing

### Initial Setup

1. **Clone the repository:**
   ```bash
   git clone https://github.com/sinouw/multilingual-video-processor.git
   cd multilingual-video-processor
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Install development tools:**
   ```bash
   make install-tools
   ```
   Or manually:
   ```bash
   go install github.com/GoogleCloudPlatform/functions-framework-go/cmd/functions-framework@latest
   go install github.com/securego/gosec/v2/cmd/gosec@latest
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   ```

4. **Set up environment variables:**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

## Development Workflow

### Running Locally

1. **Set environment variables:**
   ```bash
   export GCS_BUCKET_OUTPUT=your-test-bucket
   export GOOGLE_TRANSLATE_API_KEY=your-test-key
   ```

2. **Run the function locally:**
   ```bash
   make run-local
   ```
   Or manually:
   ```bash
   functions-framework --target=TranslateVideo --port=8080
   ```

3. **Test the endpoint:**
   ```bash
   curl http://localhost:8080/health
   ```

### Testing

1. **Run all tests:**
   ```bash
   make test
   ```

2. **Run tests with coverage:**
   ```bash
   make test-coverage
   ```

3. **Run linter:**
   ```bash
   make lint
   ```

### Building

```bash
make build
```

This creates a `function` binary in the project root.

### Code Quality Checks

Before committing:

1. **Format code:**
   ```bash
   go fmt ./...
   ```

2. **Run linter:**
   ```bash
   make lint
   ```

3. **Run tests:**
   ```bash
   make test
   ```

## Code Style Guidelines

### Go Style

- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `gofmt` for formatting (standard Go formatting)
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

### Naming Conventions

- **Packages**: Lowercase, single word when possible
- **Functions**: CamelCase, exported functions capitalized
- **Variables**: camelCase
- **Constants**: Uppercase with underscores
- **Interfaces**: End with `-er` when possible (e.g., `Reader`, `Writer`)

### Error Handling

- Always check errors
- Wrap errors with context: `fmt.Errorf("failed to do X: %w", err)`
- Return errors, don't log and ignore
- Use custom error types for specific error conditions

### Context Usage

- Always accept `context.Context` as the first parameter
- Check for context cancellation before long-running operations
- Use context timeouts for external API calls
- Pass context through the call chain

### Example

```go
func ProcessData(ctx context.Context, input string) (string, error) {
    // Check context cancellation
    select {
    case <-ctx.Done():
        return "", fmt.Errorf("operation cancelled: %w", ctx.Err())
    default:
    }

    // Perform work
    result, err := doWork(ctx, input)
    if err != nil {
        return "", fmt.Errorf("failed to process data: %w", err)
    }

    return result, nil
}
```

## Project Structure

```
multilingual-video-processor/
├── cmd/cloudfunction/     # Cloud Function entry point
├── internal/              # Internal packages (not importable)
│   ├── api/              # API handlers (health, status, webhook)
│   ├── config/           # Configuration management
│   ├── storage/          # Storage abstraction (GCS implementation)
│   ├── stt/              # Speech-to-Text module
│   ├── translation/      # Translation module
│   ├── tts/              # Text-to-Speech module
│   ├── utils/            # Utilities (retry, UUID)
│   ├── validator/        # Input validation
│   └── video/            # Video processing (FFmpeg operations)
├── pkg/models/           # Public models (importable)
├── test/                 # Test utilities and fixtures
├── examples/             # Example clients
└── docs/                 # Documentation
```

## Commit Message Conventions

Follow conventional commits format:

```
type(scope): subject

body (optional)

footer (optional)
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**
```
feat(api): add rate limiting middleware

Implements token bucket algorithm for rate limiting.
Returns 429 Too Many Requests when limit exceeded.

fix(storage): handle context cancellation in GCS operations

test(api): add tests for health endpoints
```

## Pull Request Process

1. **Create a feature branch:**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes:**
   - Write code
   - Add tests
   - Update documentation if needed

3. **Ensure quality:**
   ```bash
   make test
   make lint
   ```

4. **Commit your changes:**
   ```bash
   git add .
   git commit -m "feat(scope): description"
   ```

5. **Push and create PR:**
   ```bash
   git push origin feature/your-feature-name
   ```

6. **PR Checklist:**
   - [ ] Tests pass
   - [ ] Code is formatted
   - [ ] Linter passes
   - [ ] Documentation updated if needed
   - [ ] CHANGELOG.md updated for user-facing changes

## Debugging

### Local Debugging

1. **Add logging:**
   ```go
   slog.Debug("Debug message", "key", value)
   ```

2. **Set log level:**
   ```bash
   export LOG_LEVEL=debug
   ```

3. **Run with debugger:**
   - Use VS Code Go debugger
   - Or Delve: `dlv debug ./cmd/cloudfunction`

### Cloud Function Debugging

1. **View logs:**
   ```bash
   gcloud functions logs read multilingual-video-processor --gen2 --region=us-central1
   ```

2. **Stream logs:**
   ```bash
   gcloud functions logs tail multilingual-video-processor --gen2 --region=us-central1
   ```

## Dependencies

### Adding Dependencies

```bash
go get github.com/example/package
go mod tidy
```

### Updating Dependencies

```bash
go get -u ./...
go mod tidy
```

### Vendor Dependencies (Optional)

```bash
go mod vendor
```

## Environment Variables

See `.env.example` for all available environment variables. Required for development:

- `GCS_BUCKET_OUTPUT`: Output bucket (required)
- `GOOGLE_TRANSLATE_API_KEY`: Translation API key (required)
- `LOG_LEVEL`: Set to `debug` for verbose logging

## Common Tasks

### Add a New Language

1. Add language code to `SUPPORTED_LANGUAGES` environment variable
2. Add voice configuration in `internal/tts/voice_config.go`
3. Test translation for the new language

### Add a New API Endpoint

1. Add handler in `cmd/cloudfunction/main.go`
2. Create handler function in `internal/api/` if needed
3. Add tests
4. Update API documentation

### Modify Configuration

1. Add new config field in `internal/config/config.go`
2. Add environment variable to `.env.example`
3. Update `LoadConfig()` to read the variable
4. Update documentation

## Resources

- [Go Documentation](https://go.dev/doc/)
- [Google Cloud Functions Go Guide](https://cloud.google.com/functions/docs/concepts/go-runtime)
- [Functions Framework Go](https://github.com/GoogleCloudPlatform/functions-framework-go)
