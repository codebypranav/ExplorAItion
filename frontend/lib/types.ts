// Shapes mirror the JSON emitted by the Go backend.
// See main.go (`Out`, `WeatherInfo`) and internal/itinerary/itinerary.go (`Place`).

export type Weather = {
  temperature_c: number
  wind_kph: number
  weather_code: number
}

/** A POI as returned by POST /recommend. */
export type Place = {
  name: string
  description: string
  country: string
  score: number
  image_url: string
  latitude: number
  longitude: number
  xid: string
  /** Omitted by the backend when no rating is known. */
  rating?: number
  /** Only present when the coordinates were resolvable. */
  weather?: Weather
}

/** POST /itinerary returns one array of places per day. */
export type Itinerary = Place[][]

export type RecommendRequest = {
  query: string
  filters?: Record<string, unknown>
  top_k?: number
}

export type ItineraryRequest = {
  city: string
  days: number
  query?: string
}

/** A place plus the itinerary context it was placed in. */
export type Stop = Place & {
  /** 1-based day number. */
  day: number
  /** 1-based position within the day. */
  position: number
  /** Km from the previous stop that day, undefined for the first stop. */
  legKm?: number
}
