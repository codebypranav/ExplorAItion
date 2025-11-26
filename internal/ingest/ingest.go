package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	openTripBaseURL = "https://api.opentripmap.com/0.1/en/places"
)

// POI contains fields which we will map to Pinecone metadata, and embed.
// POI describes a Point-of-Interest response from OpenTripMap
type POI struct {
	XID         string   `json:"xid"`
	Name        string   `json:"name"`
	Kinds       string   `json:"kinds"`
	Lat         float64  `json:"lat"`
	Lon         float64  `json:"lon"`
	Country     string   `json:"country"`
	Description string   `json:"description"`
	Rate        float64  `json:"rate"`
	Image       string   `json:"image"`
	Tags        []string `json:"tags"`
}

func httpGet(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "ExplorAItion/1.0")

	client := &http.Client{Timeout: 20 * time.Second}
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

// GetCityCoords finds the latitude/longitude for a city name. Uses the geoname endpoint
func GetCityCoords(ctx context.Context, city string) (lat, lon float64, err error) {
	key := os.Getenv("OPEN_TRIP_MAP_KEY")
	if key == "" {
		return 0, 0, fmt.Errorf("OPEN_TRIP_MAP_KEY is not set")
	}
	url := fmt.Sprintf("%s/geoname?name=%s&apikey=%s", openTripBaseURL, city, key)
	b, err := httpGet(ctx, url)
	if err != nil {
		return 0, 0, err
	}
	var out struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return 0, 0, err
	}
	return out.Lat, out.Lon, nil
}

// SearchPOIsByRadius returns a list of POI references within the radius for given coords
func SearchPOIsByRadius(ctx context.Context, lat, lon float64, radiusMeters int, limit int) ([]string, error) {
	key := os.Getenv("OPEN_TRIP_MAP_KEY")
	if key == "" {
		return nil, fmt.Errorf("OPEN_TRIP_MAP_KEY is not set")
	}
	url := fmt.Sprintf("%s/radius?lon=%f&lat=%f&radius=%d&limit=%d&apikey=%s", openTripBaseURL, lon, lat, radiusMeters, limit, key)
	b, err := httpGet(ctx, url)
	if err != nil {
		return nil, err
	}
	var out struct {
		Features []struct {
			Properties struct {
				Xid string `json:"xid"`
			} `json:"properties"`
		} `json:"features"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	xids := make([]string, 0, len(out.Features))
	for _, f := range out.Features {
		xids = append(xids, f.Properties.Xid)
	}
	return xids, nil
}

// GetPOIDetails returns a POI's details for an xid
func GetPOIDetails(ctx context.Context, xid string) (POI, error) {
	key := os.Getenv("OPEN_TRIP_MAP_KEY")
	if key == "" {
		return POI{}, fmt.Errorf("OPEN_TRIP_MAP_KEY is not set")
	}
	url := fmt.Sprintf("%s/xid/%s?apikey=%s", openTripBaseURL, xid, key)
	b, err := httpGet(ctx, url)
	if err != nil {
		return POI{}, err
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return POI{}, err
	}
	// Map fields carefully since the API response can be noisy
	poi := POI{XID: xid}
	if name, ok := raw["name"].(string); ok {
		poi.Name = name
	}
	if kinds, ok := raw["kinds"].(string); ok {
		poi.Kinds = kinds
	}
	if point, ok := raw["point"].(map[string]interface{}); ok {
		if lonf, ok := point["lon"].(float64); ok {
			poi.Lon = lonf
		}
		if latf, ok := point["lat"].(float64); ok {
			poi.Lat = latf
		}
	}
	if addr, ok := raw["address"].(map[string]interface{}); ok {
		if country, ok := addr["country"].(string); ok {
			poi.Country = country
		}
	}
	if desc, ok := raw["wikipedia_extracts"].(map[string]interface{}); ok {
		if text, ok := desc["text"].(string); ok {
			poi.Description = text
		}
	} else if info, ok := raw["info"].(map[string]interface{}); ok {
		if descr, ok := info["descr"].(string); ok {
			poi.Description = descr
		}
	}
	if ratef, ok := raw["rate"].(float64); ok {
		poi.Rate = ratef
	}
	// extract preview.image
	if preview, ok := raw["preview"].(map[string]interface{}); ok {
		if img, ok := preview["source"].(string); ok {
			poi.Image = img
		}
	}
	// derive tags from kinds
	if poi.Kinds != "" {
		poi.Tags = append(poi.Tags, poi.Kinds)
	}
	return poi, nil
}

// FetchPOIsForCity: wrapper that returns detailed POIs for a city name
func FetchPOIsForCity(ctx context.Context, city string, radiusMeters int, limit int) ([]POI, error) {
	lat, lon, err := GetCityCoords(ctx, city)
	if err != nil {
		return nil, err
	}
	xids, err := SearchPOIsByRadius(ctx, lat, lon, radiusMeters, limit)
	if err != nil {
		return nil, err
	}
	pois := make([]POI, 0, len(xids))
	for _, xid := range xids {
		p, err := GetPOIDetails(ctx, xid)
		if err != nil {
			// skip noisy / missing xids rather than failing entire import
			continue
		}
		pois = append(pois, p)
	}
	return pois, nil
}
