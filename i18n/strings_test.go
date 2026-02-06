package i18n

import (
	"strings"
	"testing"
)

func TestFormatNutrition_AllSupportedLanguages_NoFmtArtifacts(t *testing.T) {
	langs := []string{"en", "es", "ru", "ua"}
	for _, lang := range langs {
		t.Run(lang, func(t *testing.T) {
			msg := FormatNutrition(1200, 1800, 40, 120, 85, 950, lang)
			if strings.Contains(msg, "%!(") {
				t.Fatalf("unexpected fmt artifact in %s message: %q", lang, msg)
			}
			if msg == "" {
				t.Fatalf("expected non-empty message for lang=%s", lang)
			}
		})
	}
}

func TestFormatNutrition_ContainsWeightValue(t *testing.T) {
	msg := FormatNutrition(100, 200, 10, 20, 30, 42.5, "en")
	if !strings.Contains(msg, "Weight") {
		t.Fatalf("expected Weight label in message: %q", msg)
	}
	if !strings.Contains(msg, "42.50g.") {
		t.Fatalf("expected weight value in message: %q", msg)
	}
}

func TestGetString_DefaultLanguageFallbackIsEnglish(t *testing.T) {
	got := GetString("unknown_command", "de")
	want := GetString("unknown_command", "en")
	if got != want {
		t.Fatalf("fallback mismatch: got %q, want %q", got, want)
	}
}
