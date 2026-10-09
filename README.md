# ExplorAItion

ExplorAItion is an AI-powered travel planning application that turns a natural-language trip request into a personalized, day-by-day itinerary. The system combines destination data, vector similarity search, and LLM-based query parsing to recommend places that match a traveler’s preferences and then organize them into a route.

## What the product does

A user can describe a trip in plain English such as:

- "3-day trip to Paris with art museums and good coffee"
- "family-friendly places in Tokyo with scenic views"
- "historic neighborhoods and food spots in Rome"

The platform then:

1. interprets the request using an LLM
2. embeds the user query and filters into a vector search flow
3. retrieves the most relevant attractions from a POI database
4. re-ranks or filters the results using metadata like country, rating, and tags
5. assembles a multi-day itinerary that is geographically and logistically coherent

## Technical architecture

### Backend
- Go
- Fiber web framework
- Pinecone vector database
- OpenAI API for embeddings and natural-language parsing
- OpenTripMap for point-of-interest source data

The backend entry point is `main.go`, which initializes:

- the Fiber app
- the Pinecone index connection
- the OpenAI client
- the route handlers for recommendation and itinerary generation

### Search and recommendation flow
The app uses a semantic retrieval pipeline:

- user input is parsed into a structured query with `internal/llm`
- the query is turned into embeddings
- the embeddings are sent to Pinecone for similarity search
- result metadata is used for filtering, ranking, and route selection

This allows the app to go beyond text matching and search by conceptual similarity.

### Itinerary generation
The itinerary engine lives in `internal/itinerary/itinerary.go` and operates on the scored POI results returned from Pinecone. It groups places into daily chunks, keeps the route geographically reasonable, and produces a structured day-by-day plan.

The logic follows a nearest-neighbor style approach: it uses the retrieved POIs, their coordinates, and sorted ranking signals to build a route that keeps nearby destinations together instead of producing a chaotic list.

### Data ingestion pipeline
The ingestion path is:

- OpenTripMap fetches POI metadata
- `internal/ingest` transforms and normalizes the data
- embeddings are generated for text descriptions and metadata
- data is upserted into Pinecone for semantic search

The seed process is triggered by:

```bash
SEED_INDEX=true go run main.go
```

## Repository structure

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
scripts/
```

## Key technical components

### `main.go`
This is the server entry point. It wires together the API, database connection, and the recommender/itinerary routes.

### `internal/llm`
This layer translates natural language requests into structured query parameters and Pinecone-friendly filters.

### `internal/embeddings`
This layer creates vector embeddings for the search index and incoming queries.

### `internal/itinerary`
This layer assembles POIs into day-by-day plans using the original match scores and geographic placement.

### `frontend/`
The frontend is a Next.js app with pages for:

- search
- itinerary generation
- interactive map rendering

## Environment variables

```bash
export PINECONE_API_KEY="your-pinecone-key"
export PINECONE_INDEX_NAME="destinations"
export PINECONE_ENVIRONMENT="us-east-1-aws"
export OPENAI_API_KEY="your-openai-key"
export OPEN_TRIP_MAP_KEY="your-opentripmap-key"

# optional
export GOOGLE_PLACES_API_KEY="your-google-places-key"
```

## Running the project

### Backend

```bash
go run main.go
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

Then open:

```text
http://localhost:3000
```

## API endpoints

### `POST /recommend`
Used to retrieve a curated list of relevant destinations.

Example body:

```json
{
  "query": "I like hiking and museums",
  "filters": {
    "country": "France"
  },
  "top_k": 10
}
```

### `POST /itinerary`
Generates a multi-day itinerary for a chosen city and travel style.

Example body:

```json
{
  "city": "Paris",
  "days": 2,
  "query": "I like museums and local food"
}
```


ExplorAItion combines retrieval-based search with route optimization rather than relying on a single monolithic prompt. The architecture allows it to:

- match user intent semantically
- filter destinations based on metadata and preferences
- scale beyond a simple keyword search
- produce itineraries that are more travel-plausible than an unordered list of attractions

## Notes

- Pinecone metadata fields include values such as `xid`, `lat`, `lon`, `name`, `kinds`, `rate`, and `country`.
- Query filters can use MongoDB-style operators such as `$gte`, `$eq`, and `$in`.
- The app was designed primarily as a prototype for personalized travel assistance and AI-enhanced trip planning.

## License

This project is licensed under the MIT License.

