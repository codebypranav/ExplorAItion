# ExplorAItion Frontend

Simple Next.js frontend that talks to the Go backend.

How to run:

1. Install dependencies
```
cd frontend
npm install
```

2. Start the dev server
```
npm run dev
```

Use env vars in `.env` or pass `NEXT_PUBLIC_API_URL` to point at your running backend, e.g. `NEXT_PUBLIC_API_URL=http://localhost:8080`.

Pro tip: During development install the types for improved DX:
```
npm install --save-dev @types/node @types/react
```

Pages:
- /search — search box & list of recommendations
- /itinerary — generate a day-by-day itinerary from a city name
