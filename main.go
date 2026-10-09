package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"sync"
	"time"

	"sort"

	"github.com/codebypranav/exploraition/internal/cache"
	"github.com/codebypranav/exploraition/internal/embeddings"
	gp "github.com/codebypranav/exploraition/internal/googleplaces"
	ingest "github.com/codebypranav/exploraition/internal/ingest"
	itin "github.com/codebypranav/exploraition/internal/itinerary"
	seed "github.com/codebypranav/exploraition/internal/seed"
	wthr "github.com/codebypranav/exploraition/internal/weather"
	pc "github.com/codebypranav/exploraition/pinecone"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	pineconeio "github.com/pinecone-io/go-pinecone/v3/pinecone"
	"google.golang.org/protobuf/types/known/structpb"
)

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found; using system environment variables")
	}
}

// itineraryRadiusKm bounds how far a stop may sit from the city centre.
const itineraryRadiusKm = 40.0

// nearbyMatches keeps only the matches within radiusKm of the given point.
// Matches without usable coordinates are dropped, since they cannot be routed.
func nearbyMatches(matches []*pineconeio.ScoredVector, lat, lon, radiusKm float64) []*pineconeio.ScoredVector {
	out := make([]*pineconeio.ScoredVector, 0, len(matches))
	for _, m := range matches {
		if m.Vector == nil || m.Vector.Metadata == nil {
			continue
		}
		mm := m.Vector.Metadata.AsMap()
		mLat, okLat := mm["lat"].(float64)
		mLon, okLon := mm["lon"].(float64)
		if !okLat || !okLon || (mLat == 0 && mLon == 0) {
			continue
		}
		if haversineKm(lat, lon, mLat, mLon) <= radiusKm {
			out = append(out, m)
		}
	}
	return out
}

// haversineKm is the great-circle distance between two points, in kilometres.
func haversineKm(aLat, aLon, bLat, bLon float64) float64 {
	const earthRadiusKm = 6371.0
	rad := func(d float64) float64 { return d * math.Pi / 180 }
	dLat := rad(bLat - aLat)
	dLon := rad(bLon - aLon)
	h := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(rad(aLat))*math.Cos(rad(bLat))*math.Sin(dLon/2)*math.Sin(dLon/2)
	return 2 * earthRadiusKm * math.Asin(math.Min(1, math.Sqrt(h)))
}

// placeEnrichment is what Google Places told us about one POI. found is false
// when Google had nothing nearby, which is cached too so that a miss does not
// cost a billed lookup on every later search.
type placeEnrichment struct {
	rating float64
	// photoName is Google's photo resource name. The displayable URL is
	// resolved separately, because it expires sooner than this entry does.
	photoName string
	found     bool
}

var (
	// Keyed by xid: one POI maps to exactly one Google place, and unlike
	// lat/lon it is a stable, exact key.
	//
	// Google's Maps Platform terms allow Places content to be cached only
	// temporarily (30 days at the time of writing; Place IDs are exempt and
	// may be stored indefinitely). PLACE_CACHE_TTL must stay inside that
	// limit, which is also why ratings are not written back into the index
	// as permanent metadata.
	placeCache = cache.New[placeEnrichment](placeCacheTTL(), 50000)

	// Weather changes quickly and is shared by everything nearby, so it is
	// keyed by coarse coordinates rather than by place.
	weatherCache = cache.New[wthr.CurrentWeather](15*time.Minute, 20000)

	// Signed photo URLs are perishable and Google does not document their
	// lifetime, so they are cached well inside any plausible expiry and
	// re-resolved from the long-lived photo name afterwards. A URL that does
	// go stale degrades to the card's placeholder rather than a broken image.
	photoCache = cache.New[string](photoURITTL(), 50000)
)

const defaultPhotoURITTL = 6 * time.Hour

func photoURITTL() time.Duration {
	if v := os.Getenv("PHOTO_URI_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
		log.Printf("ignoring invalid PHOTO_URI_TTL %q", v)
	}
	return defaultPhotoURITTL
}

// photoURL resolves a photo resource name to a URL the browser can load
// directly. The returned URL carries no API key.
func photoURL(ctx context.Context, photoName string) string {
	if photoName == "" {
		return ""
	}
	if hit, ok := photoCache.Get(photoName); ok {
		return hit
	}
	uri, err := gp.ResolvePhotoURI(ctx, photoName)
	if err != nil {
		// Cache the failure briefly by storing an empty string, so one bad
		// photo does not retry on every search.
		photoCache.Set(photoName, "")
		return ""
	}
	photoCache.Set(photoName, uri)
	return uri
}

const (
	defaultPlaceTTL = 7 * 24 * time.Hour
	// Parallel enrichment lookups in flight at once.
	enrichConcurrency = 8
)

func placeCacheTTL() time.Duration {
	if v := os.Getenv("PLACE_CACHE_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			if d > 30*24*time.Hour {
				log.Printf("PLACE_CACHE_TTL %s exceeds the 30 day limit for cached Places content; using 30 days", d)
				return 30 * 24 * time.Hour
			}
			return d
		}
		log.Printf("ignoring invalid PLACE_CACHE_TTL %q", v)
	}
	return defaultPlaceTTL
}

// weatherKey buckets coordinates to ~1km so neighbouring POIs share a reading.
func weatherKey(lat, lon float64) string {
	return fmt.Sprintf("%.2f,%.2f", lat, lon)
}

// lookupPlace returns Google enrichment for one POI, calling the API only on a
// cache miss.
func lookupPlace(ctx context.Context, xid string, lat, lon float64) placeEnrichment {
	if xid == "" {
		// Without a stable key there is nothing to cache against.
		xid = weatherKey(lat, lon)
	}
	if hit, ok := placeCache.Get(xid); ok {
		return hit
	}
	res, err := gp.GetNearestPlaceDetails(ctx, lat, lon)
	if err != nil {
		// Cache the miss as well, so repeat searches stay free.
		miss := placeEnrichment{}
		placeCache.Set(xid, miss)
		return miss
	}
	found := placeEnrichment{rating: res.Rating, photoName: res.PhotoName, found: true}
	placeCache.Set(xid, found)
	return found
}

func lookupWeather(ctx context.Context, lat, lon float64) (wthr.CurrentWeather, bool) {
	key := weatherKey(lat, lon)
	if hit, ok := weatherCache.Get(key); ok {
		return hit, true
	}
	w, err := wthr.GetCurrentWeather(ctx, lat, lon)
	if err != nil {
		return wthr.CurrentWeather{}, false
	}
	weatherCache.Set(key, w)
	return w, true
}

func main() {
	loadEnv()
	ctx := context.Background()
	idxConn, err := pc.GetIndexConnection(ctx)
	if err != nil {
		log.Fatalf("failed to connect to Pinecone index: %v", err)
	}
	if os.Getenv("SEED_INDEX") == "true" {
		if err := seed.SeedIndex(ctx, idxConn); err != nil {
			log.Fatalf("seed failed: %v", err)
		}
	}
	app := fiber.New()
	// enable CORS for frontend dev
	// requests to the backend will come from NEXT_PUBLIC_API_URL in the frontend
	app.Use(func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		if c.Method() == "OPTIONS" {
			return c.SendStatus(fiber.StatusNoContent)
		}
		return c.Next()
	})
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("🚀 ExplorAItion is live!")
	})
	app.Post("/recommend", func(c *fiber.Ctx) error {
		var body struct {
			Query   string                 `json:"query"`
			Filters map[string]interface{} `json:"filters"`
			TopK    int                    `json:"top_k"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
		}
		body.Query = strings.TrimSpace(body.Query)
		if body.Query == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "query is required"})
		}
		if body.TopK <= 0 {
			body.TopK = 10
		}
		if body.TopK > 50 {
			body.TopK = 50
		}

		// Use the raw query and explicit filters
		filters := body.Filters

		emb, err := embeddings.GenerateEmbedding(ctx, body.Query)
		if err != nil {
			log.Printf("embedding error: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate embedding"})
		}
		if emb.Empty() {
			// Nothing in the query matched the model's vocabulary.
			return c.JSON([]fiber.Map{})
		}
		req := &pineconeio.QueryByVectorValuesRequest{
			SparseValues:    &pineconeio.SparseValues{Indices: emb.Indices, Values: emb.Values},
			TopK:            uint32(body.TopK),
			IncludeMetadata: true,
		}
		if len(filters) > 0 {
			f, err := structpb.NewStruct(filters)
			if err == nil {
				req.MetadataFilter = f
			}
		}
		resp, err := idxConn.QueryByVectorValues(ctx, req)
		if err != nil {
			log.Printf("pinecone query error: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "pinecone query failed"})
		}
		if len(resp.Matches) == 0 {
			return c.JSON([]fiber.Map{})
		}
		// Convert matches to a friendlier response
		type WeatherInfo struct {
			Temperature float64 `json:"temperature_c"`
			WindSpeed   float64 `json:"wind_kph"`
			Code        int     `json:"weather_code"`
		}
		type Out struct {
			Name        string       `json:"name"`
			Description string       `json:"description"`
			Country     string       `json:"country"`
			Score       float32      `json:"score"`
			ImageURL    string       `json:"image_url"`
			Latitude    float64      `json:"latitude"`
			Longitude   float64      `json:"longitude"`
			Xid         string       `json:"xid"`
			Rating      float64      `json:"rating,omitempty"`
			Weather     *WeatherInfo `json:"weather,omitempty"`
		}
		outs := []Out{}
		for _, m := range resp.Matches {
			var out Out
			out.Score = float32(m.Score)
			if m.Vector != nil && m.Vector.Metadata != nil {
				if mm := m.Vector.Metadata.AsMap(); mm != nil {
					if nm, ok := mm["name"].(string); ok {
						out.Name = nm
					}
					if ct, ok := mm["country"].(string); ok {
						out.Country = ct
					}
					if xid, ok := mm["xid"].(string); ok {
						out.Xid = xid
					}
					if img, ok := mm["image"].(string); ok {
						out.ImageURL = img
					}
					if lat, ok := mm["lat"].(float64); ok {
						out.Latitude = lat
					}
					if lon, ok := mm["lon"].(float64); ok {
						out.Longitude = lon
					}
					if desc, ok := mm["description"].(string); ok {
						out.Description = desc
					}
				}
			}
			outs = append(outs, out)
		}

		// Enrich in parallel: each result previously cost two sequential HTTP
		// round trips, so a top_k of 50 meant 100 of them back to back.
		hasPlacesKey := os.Getenv("GOOGLE_PLACES_API_KEY") != ""
		var wg sync.WaitGroup
		sem := make(chan struct{}, enrichConcurrency)
		for i := range outs {
			if outs[i].Latitude == 0 && outs[i].Longitude == 0 {
				continue
			}
			wg.Add(1)
			go func(o *Out) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				if hasPlacesKey {
					if e := lookupPlace(ctx, o.Xid, o.Latitude, o.Longitude); e.found {
						if e.rating != 0 {
							o.Rating = e.rating
						}
						if o.ImageURL == "" && e.photoName != "" {
							o.ImageURL = photoURL(ctx, e.photoName)
						}
					}
				}
				if w, ok := lookupWeather(ctx, o.Latitude, o.Longitude); ok {
					o.Weather = &WeatherInfo{Temperature: w.Temperature, WindSpeed: w.WindSpeed, Code: w.WeatherCode}
				}
			}(&outs[i])
		}
		wg.Wait()

		if hits, misses := placeCache.Stats(); hits+misses > 0 {
			log.Printf("places cache: %d hits, %d misses (billed lookups)", hits, misses)
		}
		// Re-rank by combining pinecone semantic score + rating (if available)
		type scoredOut struct {
			Out   Out
			Score float64
		}
		scored := []scoredOut{}
		for _, o := range outs {
			r := 0.0
			if o.Rating != 0 {
				r = float64(o.Rating) / 5.0
			}
			// combine: 70% pinecone score + 30% rating
			composite := float64(o.Score)*0.7 + r*0.3
			scored = append(scored, scoredOut{Out: o, Score: composite})
		}
		sort.SliceStable(scored, func(i, j int) bool { return scored[i].Score > scored[j].Score })
		final := []Out{}
		for _, s := range scored {
			final = append(final, s.Out)
		}
		return c.JSON(final)

	})
	app.Post("/itinerary", func(c *fiber.Ctx) error {
		var body struct {
			City  string `json:"city"`
			Days  int    `json:"days"`
			Query string `json:"query"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
		}
		body.City = strings.TrimSpace(body.City)
		body.Query = strings.TrimSpace(body.Query)
		if body.City == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "city is required"})
		}
		if body.Days <= 0 {
			body.Days = 2
		}
		if body.Days > 14 {
			body.Days = 14
		}
		if body.Query == "" {
			body.Query = "top attractions in " + body.City
		}
		// get city coords
		lat, lon, err := ingest.GetCityCoords(ctx, body.City)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "failed to locate city"})
		}
		emb, err := embeddings.GenerateEmbedding(ctx, body.Query)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "embedding failed"})
		}
		// query pinecone
		if emb.Empty() {
			return c.JSON(make([][]itin.Place, body.Days))
		}
		// Ask for well beyond what the plan needs: the index spans many cities,
		// and everything outside this one is discarded below.
		topK := body.Days * 40
		if topK > 1000 {
			topK = 1000
		}
		req := &pineconeio.QueryByVectorValuesRequest{
			SparseValues:    &pineconeio.SparseValues{Indices: emb.Indices, Values: emb.Values},
			TopK:            uint32(topK),
			IncludeMetadata: true,
		}
		res, err := idxConn.QueryByVectorValues(ctx, req)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "pinecone query failed"})
		}
		// The semantic query knows nothing about geography, so a search for
		// "churches" in Rome happily matches churches in Paris. Keep only what
		// is actually near the requested city.
		local := nearbyMatches(res.Matches, lat, lon, itineraryRadiusKm)
		if len(local) == 0 {
			return c.JSON(make([][]itin.Place, body.Days))
		}
		itinerary, err := itin.GenerateItinerary(ctx, idxConn, lat, lon, local, body.Days)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to build itinerary"})
		}
		return c.JSON(itinerary)
	})
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on :%s...", port)
	log.Fatal(app.Listen(":" + port))
}
