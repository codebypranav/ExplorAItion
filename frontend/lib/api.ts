import type {
  Itinerary,
  ItineraryRequest,
  Place,
  RecommendRequest,
} from './types'

export const API_BASE =
  process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, '') || 'http://localhost:8080'

/** Backend limits from main.go, mirrored here so the UI can validate early. */
export const LIMITS = {
  topK: { min: 1, max: 50, default: 12 },
  days: { min: 1, max: 14, default: 3 },
} as const

export class ApiError extends Error {
  readonly status: number
  /** True when the request never reached the server (offline, CORS, refused). */
  readonly offline: boolean

  constructor(message: string, status = 0, offline = false) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.offline = offline
  }
}

const TIMEOUT_MS = 45_000

async function post<T>(
  path: string,
  body: unknown,
  signal?: AbortSignal,
): Promise<T> {
  const timeout = AbortSignal.timeout(TIMEOUT_MS)
  const abort = signal ? AbortSignal.any([signal, timeout]) : timeout

  let res: Response
  try {
    res = await fetch(`${API_BASE}${path}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
      signal: abort,
    })
  } catch (err) {
    // Rethrow caller-driven cancellation untouched so callers can ignore it.
    if (signal?.aborted) throw err
    if (timeout.aborted) {
      throw new ApiError(
        `The backend did not respond within ${TIMEOUT_MS / 1000}s.`,
        0,
        true,
      )
    }
    throw new ApiError(
      `Could not reach the ExplorAItion API at ${API_BASE}. Is the Go backend running?`,
      0,
      true,
    )
  }

  if (!res.ok) {
    throw new ApiError(await errorMessage(res), res.status)
  }

  try {
    return (await res.json()) as T
  } catch {
    throw new ApiError('The API returned a response that was not valid JSON.', res.status)
  }
}

/** The backend reports failures as `{"error": "..."}`; fall back to the status text. */
async function errorMessage(res: Response): Promise<string> {
  try {
    const data = await res.json()
    if (data && typeof data.error === 'string') return data.error
  } catch {
    // fall through to the generic message
  }
  return `Request failed with status ${res.status}${res.statusText ? ` (${res.statusText})` : ''}.`
}

export async function recommend(
  req: RecommendRequest,
  signal?: AbortSignal,
): Promise<Place[]> {
  const payload = {
    query: req.query.trim(),
    filters: req.filters ?? {},
    top_k: clamp(req.top_k ?? LIMITS.topK.default, LIMITS.topK),
  }
  const data = await post<Place[] | null>('/recommend', payload, signal)
  return Array.isArray(data) ? data.filter(isUsablePlace) : []
}

export async function itinerary(
  req: ItineraryRequest,
  signal?: AbortSignal,
): Promise<Itinerary> {
  const payload = {
    city: req.city.trim(),
    days: clamp(req.days, LIMITS.days),
    query: req.query?.trim() ?? '',
  }
  const data = await post<Itinerary | null>('/itinerary', payload, signal)
  if (!Array.isArray(data)) return []
  // Days with no matches come back as null, so normalise every day to an array.
  return data.map((day) => (Array.isArray(day) ? day.filter(isUsablePlace) : []))
}

/** GET / — used by the connection indicator in the header. */
export async function health(signal?: AbortSignal): Promise<boolean> {
  try {
    const res = await fetch(`${API_BASE}/`, {
      signal: signal
        ? AbortSignal.any([signal, AbortSignal.timeout(5000)])
        : AbortSignal.timeout(5000),
    })
    return res.ok
  } catch {
    return false
  }
}

function isUsablePlace(p: Place | null | undefined): p is Place {
  return !!p && typeof p === 'object'
}

function clamp(value: number, { min, max }: { min: number; max: number }) {
  if (!Number.isFinite(value)) return min
  return Math.min(max, Math.max(min, Math.round(value)))
}

export function hasCoords(p: Pick<Place, 'latitude' | 'longitude'>): boolean {
  return (
    Number.isFinite(p.latitude) &&
    Number.isFinite(p.longitude) &&
    // The backend writes 0/0 when a POI has no coordinates in its metadata.
    !(p.latitude === 0 && p.longitude === 0)
  )
}
