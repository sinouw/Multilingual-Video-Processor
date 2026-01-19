package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sinouw/multilingual-video-processor/pkg/models"
)

// Example: Simple usage of the video translation API

func main() {
	// API endpoint (replace with your deployed function URL)
	apiURL := "https://your-function-url/v1/translate"

	// Create translation request
	req := models.TranslateRequest{
		VideoURL:        "gs://your-bucket/path/to/video.mp4",
		TargetLanguages: []string{"en", "ar"},
		SourceLanguage:  "fr", // Optional, can be empty for auto-detect
	}

	// Convert to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}

	// Send request
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	// Parse response
	var translateResp models.TranslateResponse
	if err := json.Unmarshal(body, &translateResp); err != nil {
		panic(err)
	}

	fmt.Printf("Job ID: %s\n", translateResp.JobID)
	fmt.Printf("Status: %s\n", translateResp.Status)

	// Poll for job completion
	statusURL := fmt.Sprintf("%s/v1/status/%s", apiURL[:len(apiURL)-len("/translate")], translateResp.JobID)

	for {
		statusResp, err := http.Get(statusURL)
		if err != nil {
			panic(err)
		}

		body, _ := io.ReadAll(statusResp.Body)
		statusResp.Body.Close()

		var status models.StatusResponse
		json.Unmarshal(body, &status)

		fmt.Printf("Job Status: %s\n", status.Status)

		if status.Status == models.StatusCompleted || status.Status == models.StatusFailed {
			break
		}

		time.Sleep(5 * time.Second)
	}
}
