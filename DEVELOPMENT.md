# ExplorAItion Development Guide

## Quick Start

### Using the automated helper scripts (Recommended)

The easiest way to get both servers running:

```bash
./start-dev.sh
```

This will:
- ✅ Check and setup prerequisites
- ✅ Build the Go backend if needed
- ✅ Install frontend dependencies if needed
- ✅ Start the backend server on port 8080
- ✅ Start the frontend dev server on port 3000
- ✅ Display helpful information and log file locations

**To stop both servers:**
```bash
./stop-dev.sh
```

Or press `Ctrl+C` while `start-dev.sh` is running.

---

## Manual Setup (if preferred)

### Prerequisites
- Go 1.16+ installed
- Node.js 16+ and npm installed
- `.env` file in the project root with required API keys

### Required Environment Variables
```
PINECONE_API_KEY=your_key_here
PINECONE_INDEX_NAME=destinations
PINECONE_ENVIRONMENT=us-east-1-aws
OPENAI_API_KEY=your_key_here
OPEN_TRIP_MAP_KEY=your_key_here
SEED_INDEX=false
```

### Terminal 1 - Start Backend
```bash
export PATH=/usr/local/bin:$PATH
./exploraition_new
# or rebuild from source:
# go build -o exploraition_new main.go
# ./exploraition_new
```

Expected output: `listening on :8080...`

### Terminal 2 - Start Frontend
```bash
export PATH=/usr/local/bin:$PATH
npm install --prefix frontend  # (only if not already installed)
npm run dev --prefix frontend
```

Expected output: `- ready started server on 0.0.0.0:3000`

---

## Access Your Application

Once both servers are running:

- **Frontend UI**: http://localhost:3000
  - Search page: http://localhost:3000/search
  - Itinerary page: http://localhost:3000/itinerary

- **Backend API**: http://localhost:8080
  - Health check: `GET http://localhost:8080/` → "🚀 ExplorAItion is live!"
  - Recommend endpoint: `POST http://localhost:8080/recommend`
  - Itinerary endpoint: `POST http://localhost:8080/itinerary`

---

## API Endpoints

### POST /recommend
Generate recommendations based on user query and optional filters.

**Request:**
```json
{
  "query": "I like hiking",
  "filters": {"country": "France"},
  "top_k": 10
}
```

**Response:**
```json
[
  {
    "name": "Mont-Blanc",
    "country": "France",
    "description": "Highest peak in the Alps",
    "latitude": 45.8326,
    "longitude": 6.8652,
    "score": 0.95,
    "image_url": "https://...",
    "rating": 4.8,
    "weather": {
      "temperature_c": 5.2,
      "wind_kph": 12.4,
      "weather_code": 1003
    }
  }
]
```

### POST /itinerary
Generate a multi-day itinerary for a specific city.

**Request:**
```json
{
  "city": "Paris",
  "days": 3,
  "query": "I like museums"
}
```

**Response:**
```json
[
  [
    {
      "name": "Louvre Museum",
      "latitude": 48.861,
      "longitude": 2.336,
      ...
    }
  ],
  [
    {
      "name": "Musée d'Orsay",
      "latitude": 48.860,
      "longitude": 2.326,
      ...
    }
  ],
  [
    ...
  ]
]
```

---

## Troubleshooting

### Backend won't start
1. Check that Pinecone API credentials are correct in `.env`
2. Check that OpenAI API key is valid
3. Verify port 8080 is not already in use: `lsof -i :8080`
4. Review logs: `cat backend.log`

### Frontend won't start
1. Make sure Node.js is properly installed: `node --version`
2. Verify npm packages are installed: `npm list --prefix frontend`
3. Check that port 3000 is not in use: `lsof -i :3000`
4. Review logs: `cat frontend.log`

### Port already in use
If a port is already in use, you can:
- Kill the process: `./stop-dev.sh`
- Or change the port:
  - **Backend**: Set `PORT` environment variable before running
  - **Frontend**: `npm run dev --prefix frontend -- -p 3001`

### Rebuild backend
```bash
export PATH=/usr/local/bin:$PATH
go clean -cache
go build -o exploraition_new main.go
```

---

## Development Tips

### Frontend Development
The frontend uses Next.js with Hot Module Replacement (HMR), so changes to React components will hot-reload automatically.

### Backend Development
After modifying Go code, rebuild the binary:
```bash
go build -o exploraition_new main.go
./exploraition_new
```

### Testing API Endpoints
Use curl or Postman to test:
```bash
# Test backend is running
curl http://localhost:8080/

# Test recommend endpoint
curl -X POST http://localhost:8080/recommend \
  -H "Content-Type: application/json" \
  -d '{"query": "beaches", "top_k": 5}'

# Test itinerary endpoint
curl -X POST http://localhost:8080/itinerary \
  -H "Content-Type: application/json" \
  -d '{"city": "Barcelona", "days": 2, "query": "modern architecture"}'
```

---

## Project Structure

```
.
├── main.go                    # Backend entry point
├── go.mod / go.sum           # Go dependencies
├── internal/                 # Go packages
│   ├── embeddings/          # OpenAI embeddings
│   ├── ingest/              # Data ingestion
│   ├── itinerary/           # Itinerary generation
│   ├── llm/                 # LLM integration
│   ├── seed/                # Index seeding
│   ├── weather/             # Weather API
│   └── googleplaces/        # Google Places enrichment
├── pinecone/                # Pinecone client
├── frontend/                # Next.js app
│   ├── pages/              # API routes and pages
│   ├── components/         # React components
│   └── public/             # Static assets
├── start-dev.sh             # Start both servers
├── stop-dev.sh              # Stop both servers
└── README.md                # Main documentation
```

---

## Common Commands

| Command | Purpose |
|---------|---------|
| `./start-dev.sh` | Start both servers |
| `./stop-dev.sh` | Stop both servers |
| `go build -o exploraition_new main.go` | Rebuild backend |
| `npm install --prefix frontend` | Install frontend deps |
| `npm run dev --prefix frontend` | Start frontend only |
| `npm run build --prefix frontend` | Build frontend for production |
| `go mod tidy` | Clean up Go dependencies |
| `lsof -i :8080` | Check what's using port 8080 |
| `lsof -i :3000` | Check what's using port 3000 |

---

## Next Steps

1. ✅ Run `./start-dev.sh`
2. ✅ Open http://localhost:3000 in your browser
3. ✅ Try the search and itinerary features
4. ✅ Check the API endpoints with curl or Postman
5. ✅ Review logs in `backend.log` and `frontend.log` if issues occur
