package ai

import (
	"bytes"
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

// Gemini API Types
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []GeminiPart `json:"parts"`
			Role  string       `json:"role"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

func AskGemini(prompt string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY not set")
	}

	model := "gemini-2.5-flash"
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, apiKey)

	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: prompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "AI response is empty", nil
	}

	log.Printf("Gemini tokens: prompt=%d, response=%d, total=%d",
		geminiResp.UsageMetadata.PromptTokenCount,
		geminiResp.UsageMetadata.CandidatesTokenCount,
		geminiResp.UsageMetadata.TotalTokenCount)

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

// AskGeminiWithHistory sends conversation history to Google Gemini API
func AskGeminiWithHistory(history []Message) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY not set")
	}

	model := "gemini-2.5-flash"
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, apiKey)


	systemPrompt := "You are a helpful assistant as a Telegram bot(consider message max size of telegram message: 4096 char limit). Always provide concise and clear answers. Keep responses brief and to the point.\n\n"

	// Convert our Message format to Gemini format
	contents := make([]GeminiContent, 0, len(history))
	for i, msg := range history {
		role := msg.Role
		if msg.Role == "assistant" {
			role = "model"
		}

		text := msg.Content

        if i == 0 && msg.Role == "user" {
            text = systemPrompt + msg.Content
        }
		contents = append(contents, GeminiContent{
			Parts: []GeminiPart{{Text: text}},
			Role:  role,
		})
	}

	reqBody := GeminiRequest{
		Contents: contents,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "AI response is empty", nil
	}

	log.Printf("Gemini tokens: prompt=%d, response=%d, total=%d",
		geminiResp.UsageMetadata.PromptTokenCount,
		geminiResp.UsageMetadata.CandidatesTokenCount,
		geminiResp.UsageMetadata.TotalTokenCount)

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}
