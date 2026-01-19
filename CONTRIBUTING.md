# Contributing to Multilingual Video Processor

Thank you for your interest in contributing! This document provides guidelines for contributing to the project.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/your-username/multilingual-video-processor.git` (replace `your-username` with your GitHub username)
3. Create a branch: `git checkout -b feature/your-feature-name`

## Development Setup

1. Install Go 1.23 or later
2. Install dependencies: `go mod download`
3. Set up environment variables (see `.env.example`)
4. Run tests: `go test ./...`

## Code Style

- Follow Go standard formatting (`go fmt`)
- Use `golangci-lint` for linting
- Write clear, self-documenting code
- Add comments for exported functions and types
- Keep functions focused and small

## Testing

- Write unit tests for new functionality
- Maintain or improve test coverage
- Test edge cases and error conditions
- Run all tests before submitting PR: `go test ./... -v`

## Pull Request Process

1. Update documentation if needed
2. Add tests for new features
3. Ensure all tests pass
4. Update CHANGELOG.md
5. Submit PR with clear description

## Commit Messages

Use clear, descriptive commit messages:
- Start with a verb (Add, Fix, Update, Remove)
- Reference issue numbers if applicable
- Keep first line under 50 characters

Example:
```
Add support for Spanish language translation

Implements voice configuration and TTS support for Spanish (es).
Fixes #123.
```

## Questions?

Open an issue for questions or discussions.