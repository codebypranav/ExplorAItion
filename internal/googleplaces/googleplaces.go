package googleplaces

import (
	"bytes"
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

const (
	// Places API (New). The legacy maps.googleapis.com/maps/api/place endpoints
	// cannot be enabled on projects created after the legacy shutdown, so this
	// uses the current API.
	searchNearbyURL = "https://places.googleapis.com/v1/places:searchNearby"
	// Only the fields listed here are returned, and billing is per field mask
	// tier — keep this list minimal.
	fieldMask = "places.rating,places.photos"
	// Metres around the POI coordinates to look in.
	searchRadius = 50.0
	photoMaxPx   = 400
)

// GetNearestPlaceDetails finds the place at the given coordinates and returns
// its rating and a photo URL, using Google Places API (New) Nearby Search.
func GetNearestPlaceDetails(ctx context.Context, lat, lon float64) (PlaceDetails, error) {
	key := os.Getenv("GOOGLE_PLACES_API_KEY")
	if key == "" {
		return PlaceDetails{}, fmt.Errorf("GOOGLE_PLACES_API_KEY not set")
	}

	reqBody := map[string]interface{}{
		"maxResultCount": 1,
		"rankPreference": "DISTANCE",
		"locationRestriction": map[string]interface{}{
			"circle": map[string]interface{}{
				"center": map[string]float64{"latitude": lat, "longitude": lon},
				"radius": searchRadius,
			},
		},
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return PlaceDetails{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, searchNearbyURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return PlaceDetails{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", key)
	req.Header.Set("X-Goog-FieldMask", fieldMask)
	req.Header.Set("User-Agent", "ExplorAItion/1.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return PlaceDetails{}, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return PlaceDetails{}, err
	}
	if resp.StatusCode >= 300 {
		return PlaceDetails{}, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(b))
	}

	var out struct {
		Places []struct {
			Rating float64 `json:"rating"`
			Photos []struct {
				// Resource name, e.g. "places/<id>/photos/<ref>".
				Name string `json:"name"`
			} `json:"photos"`
		} `json:"places"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return PlaceDetails{}, err
	}
	if len(out.Places) == 0 {
		return PlaceDetails{}, fmt.Errorf("no place found near %f,%f", lat, lon)
	}

	p := out.Places[0]
	d := PlaceDetails{Rating: p.Rating}
	if len(p.Photos) > 0 && p.Photos[0].Name != "" {
		d.ImageURL = fmt.Sprintf(
			"https://places.googleapis.com/v1/%s/media?maxWidthPx=%d&key=%s",
			p.Photos[0].Name, photoMaxPx, key,
		)
	}
	return d, nil
}
