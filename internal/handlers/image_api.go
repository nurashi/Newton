package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type UnsplashPhoto struct {
	Urls struct {
		Regular string `json:"regular"`
	} `json:"urls"`
	User struct {
		Name string `json:"name"`
	} `json:"user"`
}

func SendUnsplashPhoto(chatID int64, query string) (string, string, error) {
	apiKey := os.Getenv("UNSPlASH_ACESS_KEY")
	apiURL := fmt.Sprintf("https://api.unsplash.com/photos/random?query=%s&client_id=%s",
		url.QueryEscape(query), apiKey)

	resp, err := http.Get(apiURL)
	if err != nil {
		return "", "",fmt.Errorf("ERROR: failed to get unsplash photo: %v", err)
	}
	defer resp.Body.Close()

	var photo UnsplashPhoto
	if err := json.NewDecoder(resp.Body).Decode(&photo); err != nil {
		return "", "", fmt.Errorf("ERROR: failed to decode unsplash photo: %v", err)
	}

	if photo.Urls.Regular == "" {
		return "", "", fmt.Errorf("ERROR: unsplash response is empty: %v", err)
	}

	caption := fmt.Sprintf("	Photo by %s", photo.User.Name)
	return photo.Urls.Regular, caption, nil
}

// Send AI generated image from Pollinations API
func SendAIImage(prompt string) (string, string, error) {
	if strings.TrimSpace(prompt) == "" {
		return "", "", fmt.Errorf("prompt cannot be empty")
	}

	imageURL := fmt.Sprintf(
		"https://image.pollinations.ai/prompt/%s",
		url.QueryEscape(strings.ReplaceAll(prompt, " ", "%20")),
	)

	caption := fmt.Sprintf("AI-generated: %s", prompt)
	return imageURL, caption, nil
}