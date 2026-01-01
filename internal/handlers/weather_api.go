package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type WheatherResponse struct {
	Location struct {
		Name      string `json:"name"`
		Country   string `json:"country"`
		Localtime string `json:"localtime"`
	} `json:"location"`

	Current struct {
		TempC     float64 `json:"temp_c"`
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
	} `json:"current"`
}

func GetWeather(city string) (string, error) {
	apiKey := os.Getenv("WHETHER_API_KEY")
	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s", apiKey, city)

	resp, err := http.Get(url)

	if err != nil {
		return "", fmt.Errorf("ERROR: failed to get resp from weather api: %v", err)
	}
	defer resp.Body.Close()

	var WheatherCity WheatherResponse

	if err := json.NewDecoder(resp.Body).Decode(&WheatherCity); err != nil {
		return "", fmt.Errorf("ERROR: failed to decode response from weather api: %v", err)
	}

	wheather := fmt.Sprintf(
		" %s, %s\n %s\n%.1f°C\n%s",
		WheatherCity.Location.Name,
		WheatherCity.Location.Country,
		WheatherCity.Current.Condition.Text,
		WheatherCity.Current.TempC,
		WheatherCity.Location.Localtime,
	)

	return wheather, nil
}
