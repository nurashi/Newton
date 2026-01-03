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
	"time"
)

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

// retryWithBackoff performs exponential backoff retry
func retryWithBackoff(maxRetries int, fn func() error) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = fn()
		if err == nil {
			return nil
		}

		if !isRetriableError(err) {
			return err
		}

		if i < maxRetries-1 {
			waitTime := time.Duration(1<<uint(i)) * time.Second
			log.Printf("Retrying after %v due to: %v", waitTime, err)
			time.Sleep(waitTime)
		}
	}
	return fmt.Errorf("max retries exceeded: %w", err)
}

func isRetriableError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return containsAny(errStr, []string{"503", "429", "UNAVAILABLE", "overloaded", "timeout"})
}

// containsAny checks if string contains any of the substrings
func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) && findSubstring(s, substr) {
			return true
		}
	}
	return false
}

// findSubstring checks if substring exists in string
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func AskGemini(prompt string) (string, error) {
	var result string
	var geminiResp GeminiResponse

	err := retryWithBackoff(4, func() error {
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return fmt.Errorf("GEMINI_API_KEY not set")
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
			return fmt.Errorf("failed to marshal request: %w", err)
		}

		req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send request: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		if resp.StatusCode != 200 {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}

		if err := json.Unmarshal(body, &geminiResp); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}

		if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
			result = "AI response is empty"
			return nil
		}

		result = geminiResp.Candidates[0].Content.Parts[0].Text
		return nil
	})

	if err != nil {
		return "", err
	}

	log.Printf("Gemini tokens: prompt=%d, response=%d, total=%d",
		geminiResp.UsageMetadata.PromptTokenCount,
		geminiResp.UsageMetadata.CandidatesTokenCount,
		geminiResp.UsageMetadata.TotalTokenCount)

	return result, nil
}

// GenerateEducationalGuide creates a comprehensive educational guide from document text
func GenerateEducationalGuide(documentText, filename, fileType string) (string, error) {
	var result string
	var geminiResp GeminiResponse

	const maxDocText = 8000
	truncated := false

	if len(documentText) > maxDocText {
		documentText = documentText[:maxDocText]
		truncated = true
	}

	truncateNote := ""
	if truncated {
		truncateNote = "\n\nNote: Document was truncated due to length. Analysis covers the first part."
	}

	prompt := fmt.Sprintf(`You are an expert educational content creator. Analyze this %s document titled "%s" and create a comprehensive EDUCATIONAL GUIDE.

DOCUMENT CONTENT:
%s
%s

YOUR TASK - Create an Educational Guide with these sections:

## Overview
Brief summary of what this document is about (2-3 sentences)

## Key Concepts & Definitions
List and explain the main concepts, terms, and definitions found in the document. Format as:
• **Term/Concept**: Clear explanation

## Main Topics Covered
Organized list of the main topics/sections with brief descriptions

## Key Takeaways
The most important points a learner should remember (numbered list)

## Study Questions
Generate 3-5 questions that would help test understanding of this material

## How Topics Connect
Explain how the different concepts in this document relate to each other

## Quick Summary for Review
A concise recap (3-5 bullet points) perfect for quick revision

Format everything in clean Markdown for Telegram. Be educational, clear, and helpful!`, fileType, filename, documentText, truncateNote)

	err := retryWithBackoff(4, func() error {
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return fmt.Errorf("GEMINI_API_KEY not set")
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
			return fmt.Errorf("failed to marshal request: %w", err)
		}

		req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send request: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		if resp.StatusCode != 200 {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}

		if err := json.Unmarshal(body, &geminiResp); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}

		if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
			result = "AI response is empty"
			return nil
		}

		result = geminiResp.Candidates[0].Content.Parts[0].Text
		return nil
	})

	if err != nil {
		return "", err
	}

	log.Printf("Educational Guide generated - tokens: prompt=%d, response=%d, total=%d",
		geminiResp.UsageMetadata.PromptTokenCount,
		geminiResp.UsageMetadata.CandidatesTokenCount,
		geminiResp.UsageMetadata.TotalTokenCount)

	return result, nil
}

// AskGeminiWithHistory sends conversation history to Google Gemini API
func AskGeminiWithHistory(history []Message) (string, error) {
	var result string
	var geminiResp GeminiResponse

	err := retryWithBackoff(4, func() error {
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return fmt.Errorf("GEMINI_API_KEY not set")
		}

		model := "gemini-2.5-flash"
		url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, apiKey)

		systemPrompt := "You are a assistant as a Telegram bot. Clear answers and conclusion in a simple words. Keep responses brief and to the point. Also text formatting should be for telegram message.\n\n"

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
			return fmt.Errorf("failed to marshal request: %w", err)
		}

		req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send request: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		if resp.StatusCode != 200 {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}

		if err := json.Unmarshal(body, &geminiResp); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}

		if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
			result = "AI response is empty"
			return nil
		}

		result = geminiResp.Candidates[0].Content.Parts[0].Text
		return nil
	})

	if err != nil {
		return "", err
	}

	log.Printf("Gemini tokens: prompt=%d, response=%d, total=%d",
		geminiResp.UsageMetadata.PromptTokenCount,
		geminiResp.UsageMetadata.CandidatesTokenCount,
		geminiResp.UsageMetadata.TotalTokenCount)

	return result, nil
}
