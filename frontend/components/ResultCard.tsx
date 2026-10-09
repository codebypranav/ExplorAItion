"use client"

import React, { useState } from 'react'

import { hasCoords } from '../lib/api'
import { formatDistance } from '../lib/geo'
import type { Place } from '../lib/types'
import WeatherChip from './WeatherChip'

type Props = {
  place: Place
  /** Rendered in a pin-style badge before the title (itinerary stop number). */
  index?: number
  /** Km from the previous stop, shown as a leg hint. */
  legKm?: number
  active?: boolean
  onSelect?: (place: Place) => void
}

export default function ResultCard({
  place,
  index,
  legKm,
  active = false,
  onSelect,
}: Props) {
  const mappable = hasCoords(place)
  const country = place.country?.trim()
  // Third-party photo hosts 404 often enough to need a fallback.
  const [imageFailed, setImageFailed] = useState(false)

  return (
    <li
      className="place-card"
      data-active={active || undefined}
      onClick={() => onSelect?.(place)}
      style={onSelect ? { cursor: 'pointer' } : undefined}
    >
      <div className="place-thumb">
        {place.image_url && !imageFailed ? (
          // The backend returns arbitrary third-party URLs (OpenTripMap /
          // Google Places), so a plain img avoids next/image host config.
          // eslint-disable-next-line @next/next/no-img-element
          <img
            src={place.image_url}
            alt={place.name ? `Photo of ${place.name}` : 'Place photo'}
            loading="lazy"
            onError={() => setImageFailed(true)}
          />
        ) : (
          <span aria-hidden="true">🏛️</span>
        )}
      </div>

      <div style={{ minWidth: 0 }}>
        <div className="row" style={{ gap: '0.5rem', alignItems: 'center' }}>
          {index !== undefined && (
            <span className="stop-index" aria-hidden="true">
              {index}
            </span>
          )}
          <h3 className="place-title">{place.name || 'Unnamed place'}</h3>
        </div>

        {country && (
          <div className="small muted" style={{ marginTop: 2 }}>
            {country}
            {legKm !== undefined ? ` · ${formatDistance(legKm)} from previous stop` : ''}
          </div>
        )}

        {place.description && <p className="place-desc">{place.description}</p>}

        <div className="place-meta">
          {Number.isFinite(place.score) && (
            <span className="chip chip-score" title="Pinecone similarity score blended with rating">
              match {place.score.toFixed(3)}
            </span>
          )}
          {place.rating ? (
            <span className="chip chip-rating" title="Google Places rating">
              ★ {place.rating.toFixed(1)}
            </span>
          ) : null}
          {place.weather ? <WeatherChip weather={place.weather} /> : null}

          {mappable && onSelect && (
            <button
              type="button"
              className="btn btn-sm no-print"
              onClick={(e) => {
                e.stopPropagation()
                onSelect(place)
              }}
            >
              Show on map
            </button>
          )}
          {mappable && (
            <a
              className="chip no-print"
              href={`https://www.openstreetmap.org/?mlat=${place.latitude}&mlon=${place.longitude}#map=16/${place.latitude}/${place.longitude}`}
              target="_blank"
              rel="noreferrer noopener"
              onClick={(e) => e.stopPropagation()}
            >
              Open in OSM ↗
            </a>
          )}
        </div>
      </div>
    </li>
  )
}
