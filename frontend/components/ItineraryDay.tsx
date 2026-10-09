"use client"

import React from 'react'

import { hasCoords } from '../lib/api'
import { dayDistanceKm, formatDistance, haversineKm } from '../lib/geo'
import type { Place } from '../lib/types'
import EmptyState from './EmptyState'
import ResultCard from './ResultCard'

type Props = {
  day: Place[]
  /** 1-based day number. */
  dayNumber: number
  activeId?: string | null
  onSelect?: (place: Place) => void
}

export default function ItineraryDay({
  day,
  dayNumber,
  activeId = null,
  onSelect,
}: Props) {
  const total = dayDistanceKm(day)

  return (
    <section className="day-block" aria-labelledby={`day-${dayNumber}`}>
      <div className="day-head">
        <h3 id={`day-${dayNumber}`}>Day {dayNumber}</h3>
        <span className="small muted mono">
          {day.length} stop{day.length === 1 ? '' : 's'}
          {total > 0 ? ` · ${formatDistance(total)} total` : ''}
        </span>
      </div>

      {day.length === 0 ? (
        <EmptyState emoji="🌤️" title="A free day">
          The search did not return enough places to fill this day. Try broader
          interests or fewer days.
        </EmptyState>
      ) : (
        <ul className="place-list">
          {day.map((place, i) => {
            const prev = i > 0 ? day[i - 1] : undefined
            const legKm =
              prev && hasCoords(prev) && hasCoords(place)
                ? haversineKm(
                    prev.latitude,
                    prev.longitude,
                    place.latitude,
                    place.longitude,
                  )
                : undefined

            return (
              <React.Fragment key={place.xid || `${dayNumber}-${i}`}>
                {legKm !== undefined && (
                  <li className="leg" aria-hidden="true">
                    ↓ {formatDistance(legKm)}
                  </li>
                )}
                <ResultCard
                  place={place}
                  index={i + 1}
                  active={!!activeId && place.xid === activeId}
                  onSelect={onSelect}
                />
              </React.Fragment>
            )
          })}
        </ul>
      )}
    </section>
  )
}
