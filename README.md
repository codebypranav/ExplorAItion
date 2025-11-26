# ExplorAItion
A travel planning application that uses Pinecone to generate itineraries for you!

1. Put in historical trips. Include dates, schedules, order, activities.
2. Generate embeddings from this information.
3. Put in preferences for a new trip, location, whether you want to change pace compared to previous trips.
4. Access Pinecone vector DB for generating a new trip.

## Developer quickstart

Environment variables required for ingestion and run:
- `PINECONE_API_KEY` - Pinecone API key
- `PINECONE_INDEX_NAME` - Pinecone index name
- `OPENAI_API_KEY` - OpenAI API key for embeddings
- `OPEN_TRIP_MAP_KEY` - OpenTripMap API key (for ingestion)
- (optional) `GOOGLE_PLACES_API_KEY` - Google Places API key for enrichment

Seed the index with real travel data (OpenTripMap) by enabling:
```
SEED_INDEX=true
```

HTTP endpoints:
- `POST /recommend` -- Request body: `{ "query": "I like hiking", "filters": {"country": "France"}, "top_k": 10 }`
- `POST /itinerary` -- Request body: `{ "city": "Paris", "days": 2, "query": "I like museums" }`

Frontend (Next.js):
- `frontend/` contains a minimal Next.js app with a /search and /itinerary page.
- Run the frontend:
```
cd frontend
npm install
npm run dev
```

The ingester pulls Points of Interest (POIs) from OpenTripMap, generates embeddings using text-embedding-3-small, and upserts vectors into Pinecone with metadata such as name, country, xid, kinds (tags), lat/lon, image URL and rating (when available).

Filtering: you can include a metadata filter in the `filters` body for `/recommend`, for example filtering by `country`, `kinds`, or `rate`.

