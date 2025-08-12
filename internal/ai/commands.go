package ai

import (
	"fmt"

)

func GeneratePitch(idea string) (string, error) {
	if idea == "" {
		return "", fmt.Errorf("no idea provided")
	}

	prompt := fmt.Sprintf(`You are a startup mentor. Create a short pitch deck for the idea: "%s".
Include:
1. Elevator pitch (1-2 sentences)
2. Problem
3. Solution
4. Target audience
5. Business model`, idea)

	return Ask(prompt)
}