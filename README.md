# Multilingual Video Processor

A standalone, open-source, cloud-function-ready service that provides video translation functionality. The service takes video files as input and generates translated versions in multiple target languages using Speech-to-Text (STT), Translation, and Text-to-Speech (TTS) services.

## Features

- **Video Translation**: Translate video audio to multiple target languages
- **Speech-to-Text**: Transcribe audio from videos using Google Cloud Speech-to-Text
- **Multi-language Support**: Translate to multiple languages concurrently
- **Text-to-Speech**: Generate natural-sounding speech in target languages
- **Audio Sync**: Automatically sync translated audio with original video
- **Cloud Function Ready**: Deploy as Google Cloud Function
- **Secure**: Input validation, rate limiting, and secure credential handling
- **Observable**: Structured logging and progress tracking

## Architecture

```
STT → Translation → TTS → Audio Sync → Output
```

1. **Speech-to-Text**: Extract audio from video and transcribe to text
2. **Translation**: Translate transcribed text to target languages
3. **Text-to-Speech**: Generate audio from translated text
4. **Audio Sync**: Replace audio track in video with translated audio
5. **Output**: Upload translated videos to cloud storage

## Quick Start

### Prerequisites

- Go 1.23 or later
- Google Cloud Project with the following APIs enabled:
  - Cloud Speech-to-Text API
  - Cloud Translation API
  - Cloud Text-to-Speech API
  - Cloud Storage API
  - See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) for detailed API enablement instructions
- **FFmpeg installed** (for video processing):
  - **macOS**: `brew install ffmpeg`
  - **Linux**: `apt-get install ffmpeg` or `yum install ffmpeg`
  - **Windows**: Download from [ffmpeg.org](https://ffmpeg.org/download.html)
- **Google Cloud credentials**: Service account JSON with required permissions:
  - `roles/storage.objectAdmin` for GCS operations
  - `roles/speech.client` for Speech-to-Text API
  - `roles/cloudtranslate.user` for Translation API
  - `roles/cloudtts.user` for Text-to-Speech API

### Installation

1. Clone the repository:
```bash
git clone https://github.com/sinouw/multilingual-video-processor.git
cd multilingual-video-processor
```

2. Install dependencies:
```bash
go mod download
```

3. Configure environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. Set up Google Cloud credentials:
```bash
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account-key.json
```

## Configuration

All configuration is done via environment variables. See `.env.example` for a complete template.

### Required Environment Variables

- `GCS_BUCKET_OUTPUT`: Output bucket for translated videos (required)
- `GOOGLE_TRANSLATE_API_KEY`: Google Translation API key (required)

### Optional Environment Variables

- `GOOGLE_APPLICATION_CREDENTIALS`: Path to service account JSON (optional, can use default credentials)
- `GCS_BUCKET_INPUT`: Input bucket for GCS URLs (optional)
- `SUPPORTED_LANGUAGES`: Comma-separated list of supported languages (default: "en,ar,de,ru")
- `SOURCE_LANGUAGE`: Default source language (optional, auto-detect if empty)
- `MAX_VIDEO_DURATION`: Maximum video duration in seconds (default: 600)
- `MAX_VIDEO_SIZE_MB`: Maximum video size in MB (default: 500)
- `MAX_CONCURRENT_JOBS`: Maximum concurrent jobs (default: 10)
- `MAX_CONCURRENT_TRANSLATIONS`: Maximum concurrent translations per job (default: 3)
- `REQUEST_TIMEOUT`: Request timeout in seconds (default: 540)
- `LOG_LEVEL`: Logging level - debug, info, warn, error (default: "info")
- `API_VERSION`: API version (default: "v1")
- `ENABLE_HEALTH_CHECK`: Enable health check endpoints (default: "true")
- `RATE_LIMIT_RPM`: Rate limit requests per minute (default: 60)
- `WEBHOOK_URL`: Webhook URL for job completion notifications (optional)
- `CORS_ORIGINS`: Comma-separated CORS origins (default: "*")
- `JOB_TTL`: Job time-to-live duration (default: "24h")
- `MAX_REQUEST_BODY_SIZE_BYTES`: Maximum request body size in bytes (default: 1048576)

## API Usage

### Submit Translation Job

```bash
curl -X POST https://your-function-url/v1/translate \
  -H "Content-Type: application/json" \
  -d '{
    "videoUrl": "gs://bucket/path/to/video.mp4",
    "targetLanguages": ["en", "ar", "de"],
    "sourceLanguage": "fr"
  }'
```

### Check Job Status

```bash
curl https://your-function-url/v1/status/{jobId}
```

**Example Response:**
```json
{
  "jobId": "550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "results": {
    "en": {
      "status": "completed",
      "videoUrl": "gs://bucket/translations/job-id/en.mp4",
      "translatedText": "Hello, this is the translated text.",
      "progress": 100,
      "processedAt": "2026-01-19T12:00:00Z"
    },
    "ar": {
      "status": "completed",
      "videoUrl": "gs://bucket/translations/job-id/ar.mp4",
      "translatedText": "مرحبا، هذا هو النص المترجم.",
      "progress": 100,
      "processedAt": "2026-01-19T12:00:00Z"
    }
  }
}
```

### Health Check

```bash
curl https://your-function-url/health
```

### Webhook Configuration

Configure webhooks to receive notifications when translation jobs complete or fail. Set the `WEBHOOK_URL` environment variable to your webhook endpoint.

**Webhook Payload:**
```json
{
  "event": "job.completed",
  "jobId": "550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "results": {
    "en": {
      "status": "completed",
      "videoUrl": "gs://bucket/translations/job-id/en.mp4",
      "translatedText": "Hello, this is the translated text.",
      "progress": 100,
      "processedAt": "2026-01-19T12:00:00Z"
    }
  },
  "timestamp": "2026-01-19T12:00:00Z"
}
```

**Event Types:**
- `job.processing`: Job started processing
- `job.completed`: Job completed successfully
- `job.failed`: Job failed (includes error message in payload)

Webhooks are triggered asynchronously and include retry logic for failed deliveries.

## Deployment

### Deploy to Google Cloud Functions

1. Build and deploy:
```bash
./deploy.sh
```

2. Or use gcloud directly:
```bash
gcloud functions deploy multilingual-video-processor \
  --gen2 \
  --runtime go123 \
  --trigger-http \
  --memory 4096MB \
  --timeout 540s \
  --allow-unauthenticated
```

See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) for detailed deployment instructions.

## Supported Languages

Currently supported target languages:
- English (en)
- Arabic (ar)
- German (de)
- Russian (ru)

More languages can be added via configuration by updating the `SUPPORTED_LANGUAGES` environment variable.

## Video Format Requirements

Supported video formats:
- MP4
- AVI
- MOV
- MKV

Video size and duration limits are configurable via environment variables:
- `MAX_VIDEO_SIZE_MB`: Maximum video size in MB (default: 500)
- `MAX_VIDEO_DURATION`: Maximum video duration in seconds (default: 600)

See [docs/API.md](docs/API.md) for more details on video format requirements.

## Development

### Local Development

1. Install dependencies:
```bash
go mod download
```

2. Install Functions Framework:
```bash
go install github.com/GoogleCloudPlatform/functions-framework-go/cmd/functions-framework@latest
```

3. Run tests:
```bash
go test ./...
```

4. Run locally:
```bash
functions-framework --target=TranslateVideo --port=8080
```

The function will be available at `http://localhost:8080`. You can change the port by setting the `PORT` environment variable (defaults to 8080).

5. Test locally using the example clients:
   - See [examples/simple/main.go](examples/simple/main.go) for basic usage
   - See [examples/advanced/main.go](examples/advanced/main.go) for advanced usage with retries and polling

**Using Makefile:**

```bash
make test          # Run tests
make test-coverage # Run tests with coverage
make lint          # Run linter
make build         # Build binary
make run-local     # Run locally with Functions Framework
```

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for more development details.

### Project Structure

```
multilingual-video-processor/
├── cmd/cloudfunction/     # Cloud Function entry point
├── internal/              # Internal packages
│   ├── stt/              # Speech-to-Text module
│   ├── translation/      # Translation module
│   ├── tts/              # Text-to-Speech module
│   ├── video/            # Video processing
│   ├── storage/          # Storage abstraction
│   ├── config/           # Configuration
│   ├── validator/        # Input validation
│   ├── api/              # API handlers
│   └── utils/            # Utilities
├── pkg/models/           # Public models
├── test/                 # Tests
├── examples/             # Usage examples
└── docs/                 # Documentation
```

## Examples

The repository includes example clients demonstrating how to use the API:

- **[examples/simple/main.go](examples/simple/main.go)**: Basic usage example showing how to submit a translation job and poll for status
- **[examples/advanced/main.go](examples/advanced/main.go)**: Advanced usage with error handling, retry logic, exponential backoff, and webhook integration

Both examples demonstrate the complete workflow from job submission to completion.

## Troubleshooting

### Common Issues

**Function fails to deploy:**
- Verify all required Google Cloud APIs are enabled
- Check that the service account has necessary permissions
- Review deployment logs: `gcloud functions logs read multilingual-video-processor --gen2 --region=us-central1`

**API authentication errors:**
- Verify `GOOGLE_APPLICATION_CREDENTIALS` is set correctly or default credentials are configured
- Check that the service account has the required IAM roles
- Ensure `GOOGLE_TRANSLATE_API_KEY` is valid

**FFmpeg not found:**
- Install FFmpeg using the platform-specific commands in Prerequisites
- Verify installation: `ffmpeg -version`
- Ensure FFmpeg is in your system PATH

**Video processing failures:**
- Check video format is supported (MP4, AVI, MOV, MKV)
- Verify video size is within limits (`MAX_VIDEO_SIZE_MB`)
- Check video duration is within limits (`MAX_VIDEO_DURATION`)
- Review function logs for detailed error messages

**Timeout issues:**
- Increase `REQUEST_TIMEOUT` environment variable (default: 540 seconds)
- Consider increasing Cloud Function timeout: `--timeout=900s`
- Monitor Cloud Function metrics in Cloud Console

For more detailed troubleshooting, see [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md).

## Cost Considerations

This service uses several Google Cloud APIs that incur costs:

- **Cloud Speech-to-Text API**: Charges per minute of audio processed
- **Cloud Translation API**: Charges per character translated
- **Cloud Text-to-Speech API**: Charges per character synthesized
- **Cloud Storage API**: Charges for storage and network egress

Monitor your usage and set up billing alerts. See [Google Cloud Pricing](https://cloud.google.com/pricing) for current rates.

**Rate Limits**: The service includes rate limiting (configurable via `RATE_LIMIT_RPM`) to help manage costs and prevent API quota exhaustion.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Security

For security concerns, please see [SECURITY.md](SECURITY.md).

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- [Documentation](docs/)
- [Issue Tracker](https://github.com/sinouw/multilingual-video-processor/issues)
- [Discussions](https://github.com/sinouw/multilingual-video-processor/discussions)

## Author

**Yassine El Ouni** ([@sinouw](https://github.com/sinouw))

## Acknowledgments

Built with:
- Google Cloud Speech-to-Text API
- Google Cloud Translation API
- Google Cloud Text-to-Speech API
- FFmpeg