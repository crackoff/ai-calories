package service

import (
	"testing"
)

func TestUpdateLanguage_ValidCodes(t *testing.T) {
	valid := []string{"en", "es-419", "pt-BR", "ru", "de", "fr"}
	for _, lang := range valid {
		t.Run(lang, func(t *testing.T) {
			svc := &UserService{userRepo: nil}
			// We only want to test the validation logic; bypass the repo call
			// by checking that a ValidationError is NOT returned for valid codes.
			// The repo is nil so we expect a nil-pointer panic only if validation passes.
			// Use a recover to distinguish validation errors from repo panics.
			var validationErr *ValidationError
			func() {
				defer func() { recover() }() // absorb nil-repo panic
				err := svc.UpdateLanguage(1, lang)
				if ve, ok := err.(*ValidationError); ok {
					validationErr = ve
				}
			}()
			if validationErr != nil {
				t.Fatalf("language %q unexpectedly failed validation: %v", lang, validationErr)
			}
		})
	}
}

func TestUpdateLanguage_InvalidCode(t *testing.T) {
	invalid := []string{"", "zh", "en-US", "ES", "portuguese"}
	for _, lang := range invalid {
		t.Run(lang, func(t *testing.T) {
			svc := &UserService{userRepo: nil}
			err := svc.UpdateLanguage(1, lang)
			if err == nil {
				t.Fatalf("expected error for language %q, got nil", lang)
			}
			ve, ok := err.(*ValidationError)
			if !ok {
				t.Fatalf("expected *ValidationError, got %T: %v", err, err)
			}
			if ve.Field != "language" {
				t.Fatalf("expected field=%q, got %q", "language", ve.Field)
			}
		})
	}
}
