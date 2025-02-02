package ai

import (
	"bytes"
	"log"
	"os"
)

type AI interface {
	// QuerySimple sends a user prompt to the AI model along with a system prompt
	// and a temperature value to control the randomness of the response.
	// It returns the AI's response as a string or an error if the process fails.
	QuerySimple(system string, user string, temperature int) (string, error)
	// RecognizeImage takes an image buffer, a system prompt, and a user prompt,
	// and returns a string response from the AI, or an error if the AI fails.
	RecognizeImage(img bytes.Buffer, system string, user string) (string, error)
}

type Classifier interface {
	GetAI() AI
}

func NewClassifier(provider string, classifier string) Classifier {
	var ai AI
	switch provider {
	case "openai":
		apiKey := os.Getenv("OPENAI_TOKEN")
		model_text := os.Getenv("AI_MODEL_TEXT")
		model_image := os.Getenv("AI_MODEL_IMAGE")
		ai = NewOpenAI(apiKey, model_text, model_image)
	default:
		log.Fatalf("unknown AI provider: %s", provider)
	}

	switch classifier {
	case "food":
		return FoodClassifier{ai: ai}
	case "expenses":
		return ExpensesClassifier{ai: ai}
	}

	log.Fatalf("unknown classifier: %s", classifier)
	return nil
}
