package translation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	GoogleTranslateAPIURL = "https://translation.googleapis.com/language/translate/v2"
)

// TranslateText translates text from source language to target language using Google Cloud Translation API
func TranslateText(ctx context.Context, text string, sourceLanguage string, targetLanguage string) (string, error) {
	slog.Info("Translating text",
		"targetLanguage", targetLanguage,
		"sourceLanguage", sourceLanguage,
		"textLength", len(text))

	apiKey := os.Getenv("GOOGLE_TRANSLATE_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("Google Translate API key not configured (GOOGLE_TRANSLATE_API_KEY)")
	}

	// Prepare request
	requestURL := fmt.Sprintf("%s?key=%s", GoogleTranslateAPIURL, apiKey)
	data := url.Values{}
	data.Set("q", text)

	// Set source language - if empty, API will auto-detect
	if sourceLanguage != "" {
		data.Set("source", sourceLanguage)
	}

	data.Set("target", targetLanguage)
	data.Set("format", "text")

	req, err := http.NewRequestWithContext(ctx, "POST", requestURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send request with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		// Check if error is due to context cancellation
		if ctx.Err() != nil {
			return "", fmt.Errorf("translation cancelled: %w", ctx.Err())
		}
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Google Translate API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var googleResp GoogleTranslateResponse
	err = json.Unmarshal(body, &googleResp)
	if err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(googleResp.Data.Translations) == 0 {
		return "", fmt.Errorf("no translations returned")
	}

	translatedText := googleResp.Data.Translations[0].TranslatedText
	slog.Info("Translation completed",
		"targetLanguage", targetLanguage,
		"translatedLength", len(translatedText))

	return translatedText, nil
}

// GoogleTranslateResponse represents the response from Google Translate API
type GoogleTranslateResponse struct {
	Data struct {
		Translations []struct {
			TranslatedText         string `json:"translatedText"`
			DetectedSourceLanguage string `json:"detectedSourceLanguage,omitempty"`
		} `json:"translations"`
	} `json:"data"`
}
