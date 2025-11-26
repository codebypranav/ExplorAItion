package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// CurrentWeather contains simplified weather data
type CurrentWeather struct {
	Temperature float64 `json:"temperature_c"`
	WindSpeed   float64 `json:"wind_kph"`
	WeatherCode int     `json:"weather_code"`
}

func httpGet(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "ExplorAItion/1.0")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(b))
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// GetCurrentWeather returns a simplified weather report using Open-Meteo free API
func GetCurrentWeather(ctx context.Context, lat, lon float64) (CurrentWeather, error) {
	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&current_weather=true", lat, lon)
	b, err := httpGet(ctx, url)
	if err != nil {
		return CurrentWeather{}, err
	}
	var out struct {
		CurrentWeather struct {
			Temperature float64 `json:"temperature"`
			WindSpeed   float64 `json:"windspeed"`
			WeatherCode int     `json:"weathercode"`
		} `json:"current_weather"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return CurrentWeather{}, err
	}
	return CurrentWeather{
		Temperature: out.CurrentWeather.Temperature,
		WindSpeed:   out.CurrentWeather.WindSpeed,
		WeatherCode: out.CurrentWeather.WeatherCode,
	}, nil
}
