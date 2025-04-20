package main

import (
	"context"
	"log"
	"os"

	"github.com/codebypranav/exploraition/internal/embeddings"
	pc "github.com/codebypranav/exploraition/pinecone"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	pineconeio "github.com/pinecone-io/go-pinecone/v3/pinecone"
	openai "github.com/sashabaranov/go-openai"
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
	openaiClient := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("🚀 ExplorAItion is live!")
	})
	app.Post("/recommend", func(c *fiber.Ctx) error {
		var body struct {
			Query string `json:"query"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
		}
		emb, err := embeddings.GenerateEmbedding(ctx, openaiClient, body.Query)
		if err != nil {
			log.Printf("embedding error: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate embedding"})
		}
		req := &pineconeio.QueryByVectorValuesRequest{
			Vector:          emb,
			TopK:            10,
			IncludeMetadata: true,
		}
		resp, err := idxConn.QueryByVectorValues(ctx, req)
		if err != nil {
			log.Printf("pinecone query error: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "pinecone query failed"})
		}
		return c.JSON(resp.Matches)

	})
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on :%s...", port)
	log.Fatal(app.Listen(":" + port))
}
