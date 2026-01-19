#!/bin/bash

# Deployment script for Multilingual Video Processor Cloud Function

set -e

# Configuration
FUNCTION_NAME="multilingual-video-processor"
REGION="us-central1"
RUNTIME="go123"
MEMORY="4096MB"
TIMEOUT="540s"
ENTRY_POINT="TranslateVideo"

# Check if required environment variables are set
if [ -z "$GCS_BUCKET_OUTPUT" ]; then
    echo "Error: GCS_BUCKET_OUTPUT environment variable is not set"
    exit 1
fi

echo "Deploying Cloud Function: $FUNCTION_NAME"

# Deploy the function
gcloud functions deploy $FUNCTION_NAME \
    --gen2 \
    --runtime=$RUNTIME \
    --region=$REGION \
    --source=. \
    --entry-point=$ENTRY_POINT \
    --trigger-http \
    --allow-unauthenticated \
    --memory=$MEMORY \
    --timeout=$TIMEOUT \
    --set-env-vars GCS_BUCKET_OUTPUT=$GCS_BUCKET_OUTPUT \
    --set-env-vars GOOGLE_TRANSLATE_API_KEY=$GOOGLE_TRANSLATE_API_KEY

if [ $? -eq 0 ]; then
    echo "Deployment successful!"
    echo ""
    echo "Function URL:"
    gcloud functions describe $FUNCTION_NAME --gen2 --region=$REGION --format="value(serviceConfig.uri)"
else
    echo "Deployment failed!"
    exit 1
fi