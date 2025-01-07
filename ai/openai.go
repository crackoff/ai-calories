package ai

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/imroc/req"
)

type OpenAI struct {
	apiKey      string
	model_text  string
	model_image string
}

type MessageContent interface{}

type TextContent struct {
	MessageContent `json:"-"`
	Type           string `json:"type"`
	Text           string `json:"text"`
}

type ImageContent struct {
	MessageContent `json:"-"`
	Type           string `json:"type"`
	ImageURL       struct {
		URL string `json:"url"`
	} `json:"image_url"`
}

type ImageURL struct {
	URL string `json:"url"`
}

type Message struct {
	Role    string           `json:"role"`
	Content []MessageContent `json:"content"`
}

func NewOpenAI(apiKey string, model_text string, model_image string) OpenAI {
	return OpenAI{apiKey: apiKey, model_text: model_text, model_image: model_image}
}

func (ai OpenAI) QuerySimple(system string, user string, temperature int) (string, error) {
	payload := map[string]interface{}{
		"model": ai.model_text,
		"messages": []interface{}{
			map[string]string{"role": "user", "content": system},
			map[string]string{"role": "user", "content": user},
		},
		//"temperature": temperature,
	}

	return ai.completion(payload)
}

func (ai OpenAI) RecognizeImage(img bytes.Buffer, system string, user string) (string, error) {
	systemMessages := []MessageContent{TextContent{Type: "text", Text: system}}
	image := ImageContent{Type: "image_url", ImageURL: ImageURL{URL: "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(img.Bytes())}}
	messages := []MessageContent{TextContent{Type: "text", Text: user}, image}

	payload := map[string]interface{}{
		"model":       ai.model_image,
		"messages":    []Message{{Role: "system", Content: systemMessages}, {Role: "user", Content: messages}},
		"max_tokens":  800,
		"temperature": 0,
	}

	return ai.completion(payload)
}

func (ai OpenAI) completion(payload interface{}) (string, error) {
	headers := req.Header{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + ai.apiKey,
	}

	r, err := req.Post("https://api.openai.com/v1/chat/completions", headers, req.BodyJSON(&payload))
	if err != nil {
		return "", err
	}

	if r.Response().StatusCode != http.StatusOK {
		resp, _ := r.ToString()
		return "", fmt.Errorf("error: %s (%s)", r.Response().Status, resp)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	_ = r.ToJSON(&result)

	return result.Choices[0].Message.Content, nil
}
