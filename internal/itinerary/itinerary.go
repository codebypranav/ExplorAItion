package itinerary

import (
	"context"
	"errors"
	"math"
	"sort"

	pineconeio "github.com/pinecone-io/go-pinecone/v3/pinecone"
)

type Place struct {
	Name        string  `json:"name"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Score       float32 `json:"score"`
	Xid         string  `json:"xid"`
	ImageURL    string  `json:"image_url"`
	Description string  `json:"description"`
	Country     string  `json:"country"`
	Rating      float64 `json:"rating"`
}

// Haversine distance
func haversineDistance(aLat, aLon, bLat, bLon float64) float64 {
	const R = 6371.0 // km
	lat1 := aLat * math.Pi / 180.0
	lat2 := bLat * math.Pi / 180.0
	dlat := (bLat - aLat) * math.Pi / 180.0
	dlon := (bLon - aLon) * math.Pi / 180.0
	va := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Cos(lat1)*math.Cos(lat2)*math.Sin(dlon/2)*math.Sin(dlon/2)
	vc := 2 * math.Atan2(math.Sqrt(va), math.Sqrt(1-va))
	return R * vc
}

// Greedy nearest neighbor route builder
func routeOrder(places []Place, startLat, startLon float64) []Place {
	if len(places) == 0 {
		return places
	}
	visited := make([]bool, len(places))
	order := []Place{}
	curLat, curLon := startLat, startLon
	for len(order) < len(places) {
		// find nearest unvisited
		best := -1
		bestDist := math.MaxFloat64
		for i := 0; i < len(places); i++ {
			if visited[i] {
				continue
			}
			d := haversineDistance(curLat, curLon, places[i].Latitude, places[i].Longitude)
			if d < bestDist {
				best = i
				bestDist = d
			}
		}
		if best == -1 {
			break
		}
		visited[best] = true
		order = append(order, places[best])
		curLat, curLon = places[best].Latitude, places[best].Longitude
	}
	return order
}

// GenerateItinerary will create a basic day-by-day list of POIs from Pinecone matches
func GenerateItinerary(ctx context.Context, idxConn *pineconeio.IndexConnection, startLat, startLon float64, matches []*pineconeio.ScoredVector, days int) ([][]Place, error) {
	if days <= 0 {
		return nil, errors.New("days must be greater than 0")
	}
	if len(matches) == 0 {
		return make([][]Place, days), nil
	}

	// Convert matches to places
	places := []Place{}
	for _, m := range matches {
		if m.Vector == nil || m.Vector.Metadata == nil {
			continue
		}
		mm := m.Vector.Metadata.AsMap()
		lat := 0.0
		lon := 0.0
		if v, ok := mm["lat"].(float64); ok {
			lat = v
		}
		if v, ok := mm["lon"].(float64); ok {
			lon = v
		}
		name := ""
		if v, ok := mm["name"].(string); ok {
			name = v
		}
		xid := ""
		if v, ok := mm["xid"].(string); ok {
			xid = v
		}
		image := ""
		if v, ok := mm["image"].(string); ok {
			image = v
		}
		desc := ""
		if v, ok := mm["description"].(string); ok {
			desc = v
		}
		country := ""
		if v, ok := mm["country"].(string); ok {
			country = v
		}
		rating := 0.0
		if v, ok := mm["rate"].(float64); ok {
			rating = v
		}
		places = append(places, Place{
			Name:        name,
			Latitude:    lat,
			Longitude:   lon,
			Score:       m.Score,
			Xid:         xid,
			ImageURL:    image,
			Description: desc,
			Country:     country,
			Rating:      rating,
		})
	}
	// Sort by score descending
	sort.SliceStable(places, func(i, j int) bool {
		return places[i].Score > places[j].Score
	})
	// Take top N based on days, e.g. days*6
	count := days * 6
	if count > len(places) {
		count = len(places)
	}
	places = places[:count]
	if len(places) == 0 {
		return make([][]Place, days), nil
	}
	// Order them by route coloring
	ordered := routeOrder(places, startLat, startLon)
	// Split into days
	itinerary := make([][]Place, days)
	placesPerDay := (len(ordered) + days - 1) / days
	for i, place := range ordered {
		day := i / placesPerDay
		if day >= days {
			day = days - 1
		}
		itinerary[day] = append(itinerary[day], place)
	}
	return itinerary, nil
}
