package main

import (
	"context"
	"log"
	"os"
	"strings"

	"sort"

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
		req := &pineconeio.QueryByVectorValuesRequest{
			Vector:          emb,
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
			// Enrich with Google Places / rating
			if out.Latitude != 0 && out.Longitude != 0 {
				if os.Getenv("GOOGLE_PLACES_API_KEY") != "" {
					if gpRes, err := gp.GetNearestPlaceDetails(ctx, out.Latitude, out.Longitude); err == nil {
						if gpRes.Rating != 0 {
							out.Rating = gpRes.Rating
						}
						if out.ImageURL == "" && gpRes.ImageURL != "" {
							out.ImageURL = gpRes.ImageURL
						}
					}
				}
				// Get weather
				if w, err := wthr.GetCurrentWeather(ctx, out.Latitude, out.Longitude); err == nil {
					out.Weather = &WeatherInfo{Temperature: w.Temperature, WindSpeed: w.WindSpeed, Code: w.WeatherCode}
				}
			}
			outs = append(outs, out)
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
		req := &pineconeio.QueryByVectorValuesRequest{Vector: emb, TopK: uint32(body.Days * 8), IncludeMetadata: true}
		res, err := idxConn.QueryByVectorValues(ctx, req)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "pinecone query failed"})
		}
		if len(res.Matches) == 0 {
			return c.JSON(make([][]itin.Place, body.Days))
		}
		itinerary, err := itin.GenerateItinerary(ctx, idxConn, lat, lon, res.Matches, body.Days)
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
