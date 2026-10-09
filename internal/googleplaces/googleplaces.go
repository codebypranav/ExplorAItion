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

// PlaceDetails returns a rating and a photo reference when possible.
type PlaceDetails struct {
	Rating float64 `json:"rating,omitempty"`
	// PhotoName is the photo's resource name, "places/<id>/photos/<ref>".
	// Resolve it with ResolvePhotoURI to get a displayable URL.
	PhotoName string `json:"photo_name,omitempty"`
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
	// Resolving a photo this way returns a signed googleusercontent URL
	// instead of redirecting, which keeps the API key out of the browser.
	photoMediaFmt = "https://places.googleapis.com/v1/%s/media?maxWidthPx=%d&skipHttpRedirect=true&key=%s"
)

// GetNearestPlaceDetails finds the place at the given coordinates and returns
// its rating and photo reference, using Google Places API (New) Nearby Search.
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
	if len(p.Photos) > 0 {
		d.PhotoName = p.Photos[0].Name
	}
	return d, nil
}

// ResolvePhotoURI turns a photo resource name into a displayable URL.
//
// The media endpoint normally 302s to the image, which would require the
// browser to call it with the API key in the query string. Asking it not to
// redirect returns the target URI instead: a signed googleusercontent.com link
// that carries no key, so the key never leaves the server.
//
// Google does not document how long these URIs stay valid, so callers should
// treat them as perishable and re-resolve periodically.
func ResolvePhotoURI(ctx context.Context, photoName string) (string, error) {
	if photoName == "" {
		return "", fmt.Errorf("empty photo name")
	}
	key := os.Getenv("GOOGLE_PLACES_API_KEY")
	if key == "" {
		return "", fmt.Errorf("GOOGLE_PLACES_API_KEY not set")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf(photoMediaFmt, photoName, photoMaxPx, key), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "ExplorAItion/1.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(b))
	}

	var out struct {
		PhotoURI string `json:"photoUri"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return "", err
	}
	if out.PhotoURI == "" {
		return "", fmt.Errorf("no photoUri returned for %s", photoName)
	}
	return out.PhotoURI, nil
}
