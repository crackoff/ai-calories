package ai

import (
	"bytes"
	"log"
	"os"
)

type AI interface {
	QuerySimple(system string, user string, temperature int) (string, error)
	RecognizeImage(img bytes.Buffer, system string, user string) (string, error)
}

type Classifier struct {
	AI
}

func NewClassifier(provider string) *Classifier {
	var ai AI
	switch provider {
	case "openai":
		apiKey := os.Getenv("OPENAI_TOKEN")
		model := os.Getenv("AI_MODEL")
		ai = NewOpenAI(apiKey, model)
	default:
		log.Fatalf("unknown AI provider: %s", provider)
	}

	return &Classifier{AI: ai}
}
