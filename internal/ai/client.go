package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Request Body, Model -> model of AI like GPT-3.5 etc.
type RequestBody struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type ResponseBody struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Choice struct {
	Message Message `json:"message"`
}

type ChatResponse struct {
	Choices []Choice `json:"choices"`
}

func Ask(prompt string) (string, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	body := RequestBody{
		Model: "mistralai/mistral-7b-instruct:free",
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		log.Printf("ERROR with convertion to json: %v", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", "https://openrouter.ai/api/v1/chat/completions", strings.NewReader(string(data)))

	if err != nil {
		log.Printf("ERROR with request with context: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://github.com/nurashi/OpenRouterProject")
	req.Header.Set("X-Title", "DevPrompt")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Printf("ERROR: %v", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API error: %s", string(raw))
	}

	var parsed ResponseBody

	_ = json.Unmarshal(raw, &parsed)

	if len(parsed.Choices) == 0 {
		return "AI response is empty", nil
	}

	return parsed.Choices[0].Message.Content, nil
}

func AskWithHistory(history []Message) (string, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")

	body := RequestBody{
		Model:    "mistralai/mistral-7b-instruct:free",
		Messages: history,
	}

	data, _ := json.Marshal(body)

	req, _ := http.NewRequestWithContext(context.Background(),
		"POST",
		"https://openrouter.ai/api/v1/chat/completions",
		strings.NewReader(string(data)),
	)

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://github.com/nurashi/OpenRouterProject")
	req.Header.Set("X-Title", "DevPrompt")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API ERROR: %s", string(raw))
	}

	var parsed ResponseBody
	_ = json.Unmarshal(raw, &parsed)

	if len(parsed.Choices) == 0 {
		return "AI response is empty", nil
	}

	return parsed.Choices[0].Message.Content, nil
}

func LMStudioAPICall(prompt string) (string, error) {
	baseURL := "http://192.168.1.81:1234/v1/chat/completions"
	model := "google/gemma-3-4b"

	body := RequestBody{
		Model: model,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		log.Printf("ERROR with conversion to json: %v", err)
		return "", err
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", baseURL, strings.NewReader(string(data)))
	if err != nil {
		log.Printf("ERROR with request with context: %v", err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ERROR: %v", err)
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API error: %s", string(raw))
	}

	var parsed ResponseBody
	err = json.Unmarshal(raw, &parsed)
	if err != nil {
		log.Printf("ERROR parsing response: %v", err)
		return "", err
	}

	if len(parsed.Choices) == 0 {
		return "AI response is empty", nil
	}
	return parsed.Choices[0].Message.Content, nil
}
