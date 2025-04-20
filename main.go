package main

import (
	"log"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on real env vars")
	}

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ExplorAItion is alive!")
	})

	log.Fatal(app.Listen(":8080"))

	pClient, err := pinecone.NewClient()
	if err != nil { log.Fatal(err) }

	index, err := pinecone.GetIndex(pClient)
	if err != nil { log.Fatal(err) }
}