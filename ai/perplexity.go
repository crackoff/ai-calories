package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func ExecutePplxQuerySimple(system string, user string, temperature int) (string, error) {
	pplxToken := os.Getenv("PPLX_TOKEN")
	client := &http.Client{}
	jsonData := fmt.Sprintf(`{"model": "llama-3-70b-instruct","messages": [{"role": "system","content": "%s"},{"role": "user","content": "%s"}],	"temperature": %d}`, system, user, temperature)

	request, err := http.NewRequest("POST", "https://api.perplexity.ai/chat/completions", bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		return "", err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", pplxToken))

	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	log.Println(string(body))

	var completion Completion
	err = json.Unmarshal(body, &completion)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return completion.Choices[0].Message.Content, nil
}

type Completion struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Created int    `json:"created"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Object  string `json:"object"`
	Choices []struct {
		Index        int    `json:"index"`
		FinishReason string `json:"finish_reason"`
		Message      struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		Delta struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}
