# API Reference

## Base URL

The API is deployed as a Google Cloud Function. The base URL will be provided after deployment.

## Endpoints

### 1. Translate Video

Translate a video from one language to multiple target languages.

**Endpoint:** `POST /v1/translate`

**Request Body:**
```json
{
  "videoUrl": "gs://bucket/path/to/video.mp4",
  "targetLanguages": ["en", "ar", "de"],
  "sourceLanguage": "fr"
}
```

**Request Parameters:**
- `videoUrl` (string, required): GCS URL (`gs://bucket/path`) or HTTPS URL of the video file
- `targetLanguages` (array, required): Array of target language codes (e.g., `["en", "ar", "de"]`)
- `sourceLanguage` (string, optional): Source language code. If not provided, will auto-detect.

**Response (202 Accepted):**
```json
{
  "jobId": "550e8400-e29b-41d4-a716-446655440000",
  "status": "processing"
}
```

**Example:**
```bash
curl -X POST https://your-function-url/v1/translate \
  -H "Content-Type: application/json" \
  -d '{
    "videoUrl": "gs://my-bucket/videos/sample.mp4",
    "targetLanguages": ["en", "ar"],
    "sourceLanguage": "fr"
  }'
```

### 2. Get Job Status

Get the status of a translation job.

**Endpoint:** `GET /v1/status/{jobId}`

**Response (200 OK):**
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

**Example:**
```bash
curl https://your-function-url/v1/status/550e8400-e29b-41d4-a716-446655440000
```

### 3. Health Check

Check if the service is healthy.

**Endpoint:** `GET /health`

**Response (200 OK):**
```json
{
  "status": "healthy",
  "timestamp": "2026-01-19T12:00:00Z",
  "version": "1.0.0"
}
```

### 4. Readiness Probe

Check if the service is ready to accept requests.

**Endpoint:** `GET /health/ready`

### 5. Liveness Probe

Check if the service is alive.

**Endpoint:** `GET /health/live`

## Status Codes

- `200 OK`: Request successful
- `202 Accepted`: Translation job submitted successfully
- `400 Bad Request`: Invalid request (missing required fields, invalid format)
- `404 Not Found`: Job not found or endpoint not found
- `500 Internal Server Error`: Server error

## Supported Languages

Currently supported target languages:
- `en` - English
- `ar` - Arabic
- `de` - German
- `ru` - Russian

Source language can be auto-detected or any valid ISO 639-1 language code.

## Error Response Format

```json
{
  "error": "Bad Request",
  "message": "videoUrl is required",
  "requestId": "550e8400-e29b-41d4-a716-446655440000"
}
```

## Rate Limits

Rate limiting can be configured via `RATE_LIMIT_RPM` environment variable (default: 60 requests per minute).

## Video Format Requirements

Supported video formats:
- MP4
- AVI
- MOV
- MKV

Maximum video duration and size can be configured via environment variables.