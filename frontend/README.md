# ExplorAItion Frontend

Next.js (App Router) client for the Go backend in the repository root.

## Running

```bash
cd frontend
npm install
npm run dev
```

Copy `.env.example` to `.env.local` to point at a backend somewhere other than
`http://localhost:8080`:

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_DEMO_MODE=false
```

### Demo mode

The app needs a seeded Pinecone index to return anything real. Set
`NEXT_PUBLIC_DEMO_MODE=true` to render a built-in Paris fixture
(`lib/sample.ts`) instead of calling the API — useful for working on the UI
before the index is ready. Sample results are always labelled as such in the
page, and the same fixture is one click away from any API error, so you can
still see the interface when the backend is down.

## Pages

| Route | What it does |
| --- | --- |
| `/` | Landing page with a prompt box that deep-links into search |
| `/search` | `POST /recommend` — ranked places, filters, sorting, map |
| `/itinerary` | `POST /itinerary` — day-by-day plan with per-day routes |

Search and itinerary state lives in the URL (`/search?q=…&k=12&country=FR`,
`/itinerary?city=Paris&days=3&q=museums`), so results can be linked and the
back button works.

## Structure

```text
app/
  layout.tsx          shared chrome (header, footer, metadata)
  page.tsx            landing page
  search/             search page + client component
  itinerary/          itinerary page + client component
components/
  Map.tsx             Leaflet map: numbered pins, day routes, dark tinting
  ResultCard.tsx      one place: photo, score, rating, weather
  ItineraryDay.tsx    one day: ordered stops with leg distances
  Header / Footer / Notices / Skeletons / EmptyState / WeatherChip
lib/
  api.ts              typed client for the Go API (timeouts, error parsing)
  types.ts            response shapes mirroring the backend structs
  geo.ts              haversine, route distance, bounds
  weather.ts          WMO weather code → label and icon
  sample.ts           demo-mode fixture
```

## Notes

- The header shows a live backend status dot; it polls `GET /` every 30s.
- Backend errors are surfaced verbatim (`{"error": "..."}`), with the API base
  URL shown when the server cannot be reached at all.
- Place photos come from arbitrary third-party hosts, so they use plain `<img>`
  rather than `next/image` (no remote host allowlist to maintain).
- Maps use OpenStreetMap raster tiles; dark mode tints the tile pane in CSS
  because the free dark basemaps now require an API key.
- Light and dark themes both follow the OS setting.
