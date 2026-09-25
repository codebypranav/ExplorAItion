# ExplorAItion

ExplorAItion is a travel planning application that uses vector search and LLM-driven intelligence to generate personalized travel itineraries based on past trips, preferences, and destination context.

## What it does

The app:
- ingests Points of Interest (POIs) from OpenTripMap
- converts them into embeddings with OpenAI
- stores them in Pinecone for similarity search
- parses natural-language trip requests into structured filters
- builds a multi-day itinerary from relevant destinations and preferences

## Architecture

- Backend: Go + Fiber
- Frontend: Next.js app router
- Vector store: Pinecone
- Embeddings: OpenAI text-embedding-3-small
- Data source: OpenTripMap

## Project structure

```text
main.go
internal/
  embeddings/
  googleplaces/
  ingest/
  itinerary/
  llm/
  seed/
  weather/
frontend/
  app/
  components/
  lib/
pinecone/
```

## Prerequisites

Before running the app, set up the following environment variables:

```bash
export PINECONE_API_KEY="your-pinecone-key"
export PINECONE_INDEX_NAME="your-index-name"
export OPENAI_API_KEY="your-openai-key"
export OPEN_TRIP_MAP_KEY="your-opentripmap-key"
# optional
export GOOGLE_PLACES_API_KEY="your-google-places-key"
```

## Run the backend

```bash
go run main.go
```

## Run the frontend

```bash
cd frontend
npm install
npm run dev
```

Then visit:

```text
http://localhost:3000
```

## Seed the Pinecone index

To ingest POI data into Pinecone:

```bash
SEED_INDEX=true go run main.go
```

## API endpoints

### POST /recommend

Body example:

```json
{
  "query": "I like hiking and museums",
  "filters": {
    "country": "France"
  },
  "top_k": 10
}
```

### POST /itinerary

Body example:

```json
{
  "city": "Paris",
  "days": 2,
  "query": "I like museums and local food"
}
```

## Notes

- The app uses a greedy nearest-neighbor style itinerary builder to plan day-by-day travel routes.
- Metadata filters can be applied to vector queries for things like `country`, `kinds`, and `rate`.
- The frontend includes search and itinerary pages plus an interactive map experience.

## License

This project is licensed under the MIT License.

