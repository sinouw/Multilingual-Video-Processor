# Deployment Guide

## Prerequisites

1. Google Cloud Project with billing enabled
2. Google Cloud SDK (`gcloud`) installed and configured
3. Required APIs enabled:
   - Cloud Functions API
   - Cloud Speech-to-Text API
   - Cloud Translation API
   - Cloud Text-to-Speech API
   - Cloud Storage API

## Enable Required APIs

```bash
gcloud services enable \
  cloudfunctions.googleapis.com \
  speech.googleapis.com \
  translate.googleapis.com \
  texttospeech.googleapis.com \
  storage.googleapis.com
```

## Setup

1. Clone the repository:
```bash
git clone https://github.com/sinouw/multilingual-video-processor.git
cd multilingual-video-processor
```

2. Configure environment variables:
```bash
export GCS_BUCKET_OUTPUT=your-output-bucket
export GOOGLE_TRANSLATE_API_KEY=your-api-key
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
```

3. Create a service account with required permissions:
```bash
gcloud iam service-accounts create video-translator \
  --display-name="Video Translator Service Account"

gcloud projects add-iam-policy-binding PROJECT_ID \
  --member="serviceAccount:video-translator@PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/storage.objectAdmin"

gcloud projects add-iam-policy-binding PROJECT_ID \
  --member="serviceAccount:video-translator@PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/speech.client"

gcloud projects add-iam-policy-binding PROJECT_ID \
  --member="serviceAccount:video-translator@PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/cloudtranslate.user"

gcloud projects add-iam-policy-binding PROJECT_ID \
  --member="serviceAccount:video-translator@PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/cloudtts.user"
```

## Deployment Methods

### Method 1: Using deploy.sh Script

```bash
./deploy.sh
```

Make sure to set required environment variables before running:
```bash
export GCS_BUCKET_OUTPUT=your-bucket
export GOOGLE_TRANSLATE_API_KEY=your-key
```

### Method 2: Using gcloud CLI

```bash
gcloud functions deploy multilingual-video-processor \
  --gen2 \
  --runtime=go123 \
  --region=us-central1 \
  --source=. \
  --entry-point=TranslateVideo \
  --trigger-http \
  --allow-unauthenticated \
  --memory=4096MB \
  --timeout=540s \
  --set-env-vars GCS_BUCKET_OUTPUT=your-bucket \
  --set-env-vars GOOGLE_TRANSLATE_API_KEY=your-key \
  --service-account=video-translator@PROJECT_ID.iam.gserviceaccount.com
```

### Method 3: Using Cloud Build

```bash
gcloud builds submit --config=cloudbuild.yaml \
  --substitutions=_GCS_BUCKET_OUTPUT=your-bucket,_GOOGLE_TRANSLATE_API_KEY=your-key
```

## Post-Deployment

1. Get the function URL:
```bash
gcloud functions describe multilingual-video-processor \
  --gen2 \
  --region=us-central1 \
  --format="value(serviceConfig.uri)"
```

2. Test the deployment:
```bash
curl https://your-function-url/health
```

## Environment Variables

Configure via `--set-env-vars` flag or Cloud Console:

- `GCS_BUCKET_OUTPUT` (required): Output bucket for translated videos
- `GOOGLE_TRANSLATE_API_KEY` (required): Google Translation API key
- `SUPPORTED_LANGUAGES`: Comma-separated list (default: en,ar,de,ru)
- `MAX_VIDEO_DURATION`: Maximum video duration in seconds (default: 600)
- `MAX_VIDEO_SIZE_MB`: Maximum video size in MB (default: 500)
- `LOG_LEVEL`: Logging level (debug, info, warn, error) (default: info)

## Troubleshooting

### Function fails to deploy

- Check that all required APIs are enabled
- Verify service account has necessary permissions
- Check logs: `gcloud functions logs read multilingual-video-processor --gen2 --region=us-central1`

### Function times out

- Increase timeout: `--timeout=900s`
- Check video file size and duration limits
- Monitor Cloud Function metrics in Cloud Console

### Translation fails

- Verify API keys are correctly set
- Check quotas for Google Cloud APIs
- Review function logs for detailed error messages