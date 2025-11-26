package googleplaces

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// PlaceDetails returns rating and image URL when possible
type PlaceDetails struct {
	Rating   float64 `json:"rating,omitempty"`
	ImageURL string  `json:"image_url,omitempty"`
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

// GetNearestPlaceDetails tries to find a place around the given coordinates using Google Places Nearby Search.
func GetNearestPlaceDetails(ctx context.Context, lat, lon float64) (PlaceDetails, error) {
	key := os.Getenv("GOOGLE_PLACES_API_KEY")
	if key == "" {
		return PlaceDetails{}, fmt.Errorf("GOOGLE_PLACES_API_KEY not set")
	}
	url := fmt.Sprintf("https://maps.googleapis.com/maps/api/place/nearbysearch/json?location=%f,%f&radius=50&key=%s", lat, lon, key)
	b, err := httpGet(ctx, url)
	if err != nil {
		return PlaceDetails{}, err
	}
	var out struct {
		Results []struct {
			Rating interface{} `json:"rating"`
			Photos []struct {
				PhotoReference string `json:"photo_reference"`
			} `json:"photos"`
		} `json:"results"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return PlaceDetails{}, err
	}
	if len(out.Results) == 0 {
		return PlaceDetails{}, fmt.Errorf("no places found")
	}
	r := out.Results[0]
	d := PlaceDetails{}
	if r.Rating != nil {
		switch v := r.Rating.(type) {
		case float64:
			d.Rating = v
		case string:
			// parse string
		}
	}
	if len(r.Photos) > 0 {
		// Build photo URL
		d.ImageURL = fmt.Sprintf("https://maps.googleapis.com/maps/api/place/photo?maxwidth=400&photoreference=%s&key=%s", r.Photos[0].PhotoReference, key)
	}
	return d, nil
}
