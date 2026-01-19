package validator

import (
	"testing"
)

func TestValidateLanguageCode(t *testing.T) {
	supported := []string{"en", "ar", "de", "ru"}

	tests := []struct {
		name      string
		language  string
		supported []string
		wantErr   bool
	}{
		{"valid language", "en", supported, false},
		{"valid language ar", "ar", supported, false},
		{"invalid language", "fr", supported, true},
		{"empty language", "", supported, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLanguageCode(tt.language, tt.supported)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLanguageCode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateLanguageCodes(t *testing.T) {
	supported := []string{"en", "ar", "de", "ru"}

	tests := []struct {
		name      string
		languages []string
		wantErr   bool
	}{
		{"valid languages", []string{"en", "ar"}, false},
		{"single language", []string{"en"}, false},
		{"empty list", []string{}, true},
		{"duplicate", []string{"en", "en"}, true},
		{"unsupported", []string{"fr"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLanguageCodes(tt.languages, supported)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLanguageCodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
