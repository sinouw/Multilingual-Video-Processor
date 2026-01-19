package validator

import (
	"fmt"
)

// ValidateLanguageCode validates a language code against supported languages
func ValidateLanguageCode(language string, supportedLanguages []string) error {
	if language == "" {
		return fmt.Errorf("language code cannot be empty")
	}

	for _, supported := range supportedLanguages {
		if language == supported {
			return nil
		}
	}

	return fmt.Errorf("unsupported language code: %s", language)
}

// ValidateLanguageCodes validates multiple language codes
func ValidateLanguageCodes(languages []string, supportedLanguages []string) error {
	if len(languages) == 0 {
		return fmt.Errorf("at least one language must be specified")
	}

	seen := make(map[string]bool)
	for _, lang := range languages {
		if seen[lang] {
			return fmt.Errorf("duplicate language code: %s", lang)
		}
		seen[lang] = true

		if err := ValidateLanguageCode(lang, supportedLanguages); err != nil {
			return err
		}
	}

	return nil
}
