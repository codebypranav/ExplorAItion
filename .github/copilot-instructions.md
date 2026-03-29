# ExplorAItion Copilot Instructions

## Project Overview
ExplorAItion is a travel planning application that generates personalized itineraries using Vector Search (Pinecone) and LLMs (OpenAI).
- **Backend**: Go (Fiber framework)
- **Frontend**: Next.js 14 (App Router)
- **Data**: OpenTripMap (POIs), Pinecone (Vector DB), OpenAI (Embeddings & Chat)

## Architecture & Data Flow

### Backend (`/`)
- **Entry Point**: `main.go` initializes the Fiber app, Pinecone connection, and OpenAI client.
- **Service Boundaries**:
  - `internal/ingest`: Fetches POIs from OpenTripMap API.
  - `internal/embeddings`: Generates vector embeddings using OpenAI `text-embedding-3-small`.
  - `internal/itinerary`: Implements the itinerary generation logic (Greedy Nearest Neighbor).
  - `internal/llm`: Parses natural language user queries into structured search filters.
  - `internal/googleplaces`: Enriches POI data (optional).
- **Data Flow**:
  1. **Ingestion**: OpenTripMap -> `ingest` -> `embeddings` -> Pinecone Upsert.
  2. **Search**: User Query -> `llm.ParseNaturalLanguageQuery` -> `embeddings` -> Pinecone Query.
  3. **Itinerary**: Search Results -> `itinerary.GenerateItinerary` -> Day-by-day plan.

### Frontend (`frontend/`)
- **Framework**: Next.js 14 with App Router (`app/` directory).
- **State Management**: React `useState` and `SWR` for data fetching.
- **Maps**: `react-leaflet` (Leaflet.js) with dynamic imports to avoid SSR issues (`components/Map.tsx`).
- **API Integration**: Calls backend endpoints (`/recommend`, `/itinerary`) via `fetch`.

## Developer Workflows

### Running the Project
- **Backend**:
  ```bash
  # Ensure .env is set (PINECONE_API_KEY, OPENAI_API_KEY, OPEN_TRIP_MAP_KEY)
  go run main.go
  ```
- **Frontend**:
  ```bash
  cd frontend
  npm run dev
  ```
- **Seeding Data**:
  To ingest new data from OpenTripMap into Pinecone:
  ```bash
  SEED_INDEX=true go run main.go
  ```

### Key Conventions & Patterns

- **Go Error Handling**: Return errors explicitly. Use `log.Printf` for server-side logging but return JSON errors to the client.
- **Pinecone Metadata**:
  - Critical fields: `xid` (OpenTripMap ID), `lat`, `lon`, `name`, `kinds` (tags), `rate` (rating), `country`.
  - Filtering: Use MongoDB-style operators (e.g., `$gte`, `$eq`) in Pinecone queries.
- **LLM Integration**:
  - Use `internal/llm` to structure unstructured user input before querying the vector DB.
  - Always fallback to raw query if LLM parsing fails.
- **Frontend Components**:
  - Use `"use client"` directive for components using hooks or browser-only APIs (like Leaflet).
  - Map components must be dynamically imported with `{ ssr: false }`.

## Critical Files
- `main.go`: Server routes and dependency injection.
- `internal/itinerary/itinerary.go`: Core logic for routing and day planning.
- `internal/ingest/ingest.go`: OpenTripMap API client.
- `frontend/app/itinerary/page.tsx`: Main UI for itinerary generation and visualization.
