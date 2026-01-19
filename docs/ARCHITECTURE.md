# Architecture

## Overview

The Multilingual Video Processor is a serverless cloud function that translates video content from one language to multiple target languages. It uses Google Cloud services for speech-to-text, translation, and text-to-speech capabilities.

## Architecture Diagram

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ HTTP POST /v1/translate
       ▼
┌─────────────────────────────────┐
│  Cloud Function (Gen2)          │
│  ┌───────────────────────────┐  │
│  │  HTTP Handler             │  │
│  │  - Request Validation     │  │
│  │  - Job Creation           │  │
│  └───────────┬───────────────┘  │
│              │                   │
│  ┌───────────▼───────────────┐  │
│  │  Translation Pipeline     │  │
│  │  1. Download Video (GCS)  │  │
│  │  2. Extract Audio (FFmpeg)│  │
│  │  3. STT (Speech API)      │  │
│  │  4. Translate (Translate) │  │
│  │  5. TTS (TTS API)         │  │
│  │  6. Sync Audio (FFmpeg)   │  │
│  │  7. Upload (GCS)          │  │
│  └───────────────────────────┘  │
└─────────────────────────────────┘
       │
       ▼
┌─────────────┐     ┌──────────────┐
│  GCS        │     │ Google APIs  │
│  (Storage)  │     │ - Speech     │
└─────────────┘     │ - Translate  │
                    │ - TTS        │
                    └──────────────┘
```

## Components

### 1. HTTP Handler (`cmd/cloudfunction/main.go`)

- Routes requests to appropriate handlers
- Validates incoming requests
- Creates translation jobs
- Returns job status

### 2. Storage Layer (`internal/storage/`)

- Abstracts storage operations
- GCS implementation for Google Cloud Storage
- Handles download/upload of video files

### 3. STT Module (`internal/stt/`)

- Extracts audio from video using FFmpeg
- Transcribes audio to text using Google Cloud Speech-to-Text
- Supports auto-detection or explicit language hints

### 4. Translation Module (`internal/translation/`)

- Translates text using Google Cloud Translation API
- Supports multiple target languages
- Handles source language auto-detection

### 5. TTS Module (`internal/tts/`)

- Generates speech from translated text
- Configurable voice per language
- Speed adjustment to match original video duration

### 6. Video Processing (`internal/video/`)

- Audio-video synchronization using FFmpeg
- Duration calculation utilities
- Video format support

### 7. Validation (`internal/validator/`)

- Request validation
- Language code validation
- URL format validation

## Data Flow

1. **Request**: Client sends video URL and target languages
2. **Validation**: Request is validated (URL format, languages, limits)
3. **Job Creation**: Unique job ID is generated and stored
4. **Video Download**: Video is downloaded from GCS to temporary storage
5. **Audio Extraction**: Audio track is extracted using FFmpeg
6. **Transcription**: Audio is transcribed to text using Speech-to-Text API
7. **Translation**: For each target language:
   - Text is translated using Translation API
   - Translated text is converted to speech using TTS API
   - New audio is synchronized with original video using FFmpeg
   - Translated video is uploaded to GCS
8. **Response**: Job status is updated and client can poll for results

## Concurrency

- Multiple target languages are processed concurrently
- Maximum concurrency is configurable via `MAX_CONCURRENT_TRANSLATIONS`
- Semaphore pattern ensures resource limits are respected

## Error Handling

- Errors are captured at each step
- Failed translations are marked but don't fail the entire job
- Detailed error messages are stored in job results
- Temporary files are cleaned up on error

## Scalability

- Serverless architecture scales automatically
- Each translation job is independent
- Stateless design allows horizontal scaling
- Job status stored in-memory (can be replaced with persistent storage)

## Security

- Input validation prevents malicious requests
- CORS configuration restricts cross-origin access
- Service account authentication for Google Cloud services
- API keys stored as environment variables
- Video file type validation

## Monitoring

- Structured logging with correlation IDs
- Health check endpoints for monitoring
- Job progress tracking
- Error categorization and reporting