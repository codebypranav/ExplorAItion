import { hasCoords } from './api'
import type { Itinerary, Place, Stop } from './types'

/** Great-circle distance in km — same formula the backend routes with. */
export function haversineKm(
  aLat: number,
  aLon: number,
  bLat: number,
  bLon: number,
): number {
  const R = 6371
  const toRad = (d: number) => (d * Math.PI) / 180
  const dLat = toRad(bLat - aLat)
  const dLon = toRad(bLon - aLon)
  const h =
    Math.sin(dLat / 2) ** 2 +
    Math.cos(toRad(aLat)) * Math.cos(toRad(bLat)) * Math.sin(dLon / 2) ** 2
  return 2 * R * Math.asin(Math.min(1, Math.sqrt(h)))
}

export function formatDistance(km: number): string {
  if (km < 1) return `${Math.round(km * 1000)} m`
  if (km < 10) return `${km.toFixed(1)} km`
  return `${Math.round(km)} km`
}

/** Flatten an itinerary into stops annotated with day, position and leg length. */
export function toStops(itin: Itinerary): Stop[] {
  const stops: Stop[] = []
  itin.forEach((day, dayIdx) => {
    day.forEach((place, i) => {
      const prev = i > 0 ? day[i - 1] : undefined
      const legKm =
        prev && hasCoords(prev) && hasCoords(place)
          ? haversineKm(prev.latitude, prev.longitude, place.latitude, place.longitude)
          : undefined
      stops.push({ ...place, day: dayIdx + 1, position: i + 1, legKm })
    })
  })
  return stops
}

/** Total walking/driving distance of one day's route. */
export function dayDistanceKm(day: Place[]): number {
  const pts = day.filter(hasCoords)
  let total = 0
  for (let i = 1; i < pts.length; i++) {
    total += haversineKm(
      pts[i - 1].latitude,
      pts[i - 1].longitude,
      pts[i].latitude,
      pts[i].longitude,
    )
  }
  return total
}

/** A rough [[south, west], [north, east]] box around the given points. */
export function bounds(
  points: Array<Pick<Place, 'latitude' | 'longitude'>>,
): [[number, number], [number, number]] | null {
  const pts = points.filter(hasCoords)
  if (pts.length === 0) return null
  const lats = pts.map((p) => p.latitude)
  const lons = pts.map((p) => p.longitude)
  return [
    [Math.min(...lats), Math.min(...lons)],
    [Math.max(...lats), Math.max(...lons)],
  ]
}
