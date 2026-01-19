package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Example: Advanced usage with error handling and webhooks

func main() {
	apiURL := "https://your-function-url/v1/translate"

	req := map[string]interface{}{
		"videoUrl":        "gs://your-bucket/video.mp4",
		"targetLanguages": []string{"en", "ar", "de", "ru"},
		"sourceLanguage":  "",
	}

	jsonData, _ := json.Marshal(req)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Send request with retry logic
	var resp *http.Response
	var err error
	maxRetries := 3

	for i := 0; i < maxRetries; i++ {
		resp, err = client.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
		if err == nil && resp.StatusCode < 500 {
			break
		}
		if i < maxRetries-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusAccepted {
		fmt.Printf("Error: %s\n", string(body))
		return
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	jobID := result["jobId"].(string)
	fmt.Printf("Job submitted: %s\n", jobID)

	// Poll with exponential backoff
	pollInterval := 2 * time.Second
	maxPollTime := 10 * time.Minute
	startTime := time.Now()

	for time.Since(startTime) < maxPollTime {
		statusURL := fmt.Sprintf("%s/v1/status/%s",
			apiURL[:len(apiURL)-len("/translate")], jobID)

		statusResp, err := client.Get(statusURL)
		if err != nil {
			fmt.Printf("Error polling: %v\n", err)
			time.Sleep(pollInterval)
			continue
		}

		body, _ := io.ReadAll(statusResp.Body)
		statusResp.Body.Close()

		var status map[string]interface{}
		json.Unmarshal(body, &status)

		fmt.Printf("Status: %s\n", status["status"])

		if status["status"] == "completed" || status["status"] == "failed" {
			results := status["results"].(map[string]interface{})
			for lang, result := range results {
				resultMap := result.(map[string]interface{})
				if resultMap["status"] == "completed" {
					fmt.Printf("  %s: %s\n", lang, resultMap["videoUrl"])
				}
			}
			break
		}

		time.Sleep(pollInterval)
		pollInterval = time.Duration(float64(pollInterval) * 1.5) // Exponential backoff
		if pollInterval > 30*time.Second {
			pollInterval = 30 * time.Second
		}
	}
}
