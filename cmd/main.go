package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

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

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file")
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		fmt.Println("OPENROUTER_API_KEY not found.")
		return
	}

	
	body := RequestBody{
		Model: "mistralai/mistral-7b-instruct:free", 
		Messages: []Message{
			{Role: "user", Content: "Capital of Kazakhstan"},
		},
	}

	data, _ := json.Marshal(body)

	req, _ := http.NewRequestWithContext(context.Background(),
		"POST",
		"https://openrouter.ai/api/v1/chat/completions",
		bytes.NewBuffer(data),
	)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://github.com/yourname/yourproject")
	req.Header.Set("X-Title", "DevPrompt")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		fmt.Printf("API Error (%d): %s\n", resp.StatusCode, string(raw))
		return
	}

	var parsed ResponseBody
	if err := json.Unmarshal(raw, &parsed); err != nil {
		fmt.Println("JSON Parse Error:", err)
		fmt.Println("Response:", string(raw))
		return
	}

	fmt.Println("Response:", parsed.Choices[0].Message.Content)
}