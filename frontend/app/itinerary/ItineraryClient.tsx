"use client"

import dynamic from 'next/dynamic'
import { useRouter, useSearchParams } from 'next/navigation'
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import EmptyState from '../../components/EmptyState'
import ItineraryDay from '../../components/ItineraryDay'
import { DemoNotice, ErrorNotice } from '../../components/Notices'
import { PlaceListSkeleton } from '../../components/Skeletons'
import { itinerary as fetchItinerary, LIMITS } from '../../lib/api'
import { dayDistanceKm, formatDistance } from '../../lib/geo'
import { sampleItinerary } from '../../lib/sample'
import type { Itinerary, Place } from '../../lib/types'

const Map = dynamic(() => import('../../components/Map'), {
  ssr: false,
  loading: () => <div className="skeleton" style={{ height: 480 }} />,
})

const DEMO_DEFAULT = process.env.NEXT_PUBLIC_DEMO_MODE === 'true'

const IDEAS = [
  'art museums and good coffee',
  'street food and night markets',
  'parks, gardens and viewpoints',
  'historic churches and old quarters',
]

export default function ItineraryClient() {
  const router = useRouter()
  const params = useSearchParams()

  const urlCity = params.get('city') ?? ''
  const urlDays = Number(params.get('days')) || LIMITS.days.default
  const urlQuery = params.get('q') ?? ''

  const [city, setCity] = useState(urlCity || 'Paris')
  const [days, setDays] = useState(urlDays)
  const [query, setQuery] = useState(urlQuery)

  const [plan, setPlan] = useState<Itinerary | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<unknown>(null)
  const [isSample, setIsSample] = useState(false)
  /** null = every day shown at once. */
  const [activeDay, setActiveDay] = useState<number | null>(null)
  const [activeId, setActiveId] = useState<string | null>(null)

  const inFlight = useRef<AbortController | null>(null)

  const loadSample = useCallback(() => {
    inFlight.current?.abort()
    setLoading(false)
    setError(null)
    setIsSample(true)
    setPlan(sampleItinerary(days))
    setActiveDay(null)
    setActiveId(null)
  }, [days])

  const generate = useCallback(
    async (cityName: string, dayCount: number, interests: string) => {
      inFlight.current?.abort()
      const controller = new AbortController()
      inFlight.current = controller

      setLoading(true)
      setError(null)
      setIsSample(false)
      setActiveId(null)
      setActiveDay(null)

      try {
        const result = await fetchItinerary(
          { city: cityName, days: dayCount, query: interests },
          controller.signal,
        )
        setPlan(result)
      } catch (err) {
        if (controller.signal.aborted) return
        setError(err)
        setPlan(null)
      } finally {
        if (!controller.signal.aborted) setLoading(false)
      }
    },
    [],
  )

  useEffect(() => {
    if (!urlCity.trim()) {
      setPlan(null)
      return
    }
    if (DEMO_DEFAULT) {
      setIsSample(true)
      setPlan(sampleItinerary(urlDays))
      return
    }
    generate(urlCity, urlDays, urlQuery)
  }, [urlCity, urlDays, urlQuery, generate])

  useEffect(() => () => inFlight.current?.abort(), [])

  const submit = () => {
    const trimmed = city.trim()
    if (!trimmed) return
    const next = new URLSearchParams({ city: trimmed, days: String(days) })
    if (query.trim()) next.set('q', query.trim())

    if (next.toString() === params.toString()) {
      // Same URL, so the effect below will not re-run — regenerate directly.
      generate(trimmed, days, query)
      return
    }
    router.push(`/itinerary?${next.toString()}`)
  }

  const visibleDays = useMemo(() => {
    if (!plan) return []
    return plan
      .map((stops, i) => ({ dayNumber: i + 1, stops }))
      .filter((d) => activeDay === null || d.dayNumber === activeDay)
  }, [plan, activeDay])

  // Pins are numbered within their day; one route is drawn per visible day.
  const mapPoints = useMemo(
    () =>
      visibleDays.flatMap(({ dayNumber, stops }) =>
        stops.map((p, i) => ({
          id: p.xid || `${dayNumber}-${i}`,
          name: p.name,
          latitude: p.latitude,
          longitude: p.longitude,
          description: p.description,
          rating: p.rating,
          weather: p.weather,
          label: i + 1,
        })),
      ),
    [visibleDays],
  )

  const totals = useMemo(() => {
    if (!plan) return null
    const stops = plan.flat().length
    const km = plan.reduce((sum, day) => sum + dayDistanceKm(day), 0)
    return { stops, km }
  }, [plan])

  const select = (place: Place) => setActiveId(place.xid || null)

  return (
    <div className="shell page stack-lg">
      <header>
        <p className="eyebrow">Day-by-day planning</p>
        <h1 style={{ marginBottom: '0.3rem' }}>Build an itinerary</h1>
        <p style={{ maxWidth: '62ch' }}>
          Pick a city and how long you have. Matching places are grouped into
          days and ordered so each day stays in one part of town.
        </p>
      </header>

      <form
        className="panel stack"
        onSubmit={(e) => {
          e.preventDefault()
          submit()
        }}
      >
        <div
          className="row"
          style={{ gap: '1rem', alignItems: 'flex-end', flexWrap: 'wrap' }}
        >
          <div className="field" style={{ flex: '1 1 200px' }}>
            <label htmlFor="city">City</label>
            <input
              id="city"
              value={city}
              onChange={(e) => setCity(e.target.value)}
              placeholder="Paris"
              required
            />
          </div>
          <div className="field" style={{ width: 110 }}>
            <label htmlFor="days">Days</label>
            <input
              id="days"
              type="number"
              min={LIMITS.days.min}
              max={LIMITS.days.max}
              value={days}
              onChange={(e) => setDays(Number(e.target.value))}
            />
          </div>
          <div className="field" style={{ flex: '2 1 280px' }}>
            <label htmlFor="interests">Interests (optional)</label>
            <input
              id="interests"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="museums, markets, rooftop views"
            />
          </div>
          <button type="submit" className="btn btn-primary" disabled={loading || !city.trim()}>
            {loading ? 'Planning…' : 'Generate plan'}
          </button>
        </div>

        <div className="row">
          <span className="eyebrow">Ideas</span>
          {IDEAS.map((idea) => (
            <button
              key={idea}
              type="button"
              className="chip"
              onClick={() => setQuery(idea)}
            >
              {idea}
            </button>
          ))}
        </div>
      </form>

      {isSample && (
        <DemoNotice onExit={() => generate(urlCity || 'Paris', days, query)} />
      )}
      {error ? (
        <ErrorNotice
          error={error}
          onRetry={() => generate(urlCity, urlDays, urlQuery)}
          onUseSample={loadSample}
        />
      ) : null}

      {loading && <PlaceListSkeleton count={3} />}

      {!loading && !error && plan === null && (
        <EmptyState emoji="🗺️" title="No plan yet">
          Fill in a city above to generate a routed itinerary.
        </EmptyState>
      )}

      {!loading && plan && plan.flat().length === 0 && (
        <EmptyState emoji="🫙" title="No places came back">
          The index had no matches for this city and interest. Check that the
          Pinecone index is seeded for this destination.
        </EmptyState>
      )}

      {!loading && plan && plan.flat().length > 0 && (
        <>
          <div className="row-between">
            <div className="row">
              <span className="small muted">
                <strong className="mono">{totals?.stops}</strong> stops across{' '}
                <strong className="mono">{plan.length}</strong> day
                {plan.length === 1 ? '' : 's'}
                {totals && totals.km > 0
                  ? ` · ${formatDistance(totals.km)} of travel`
                  : ''}
              </span>
            </div>
            <button
              type="button"
              className="btn btn-sm no-print"
              onClick={() => window.print()}
            >
              Print / save as PDF
            </button>
          </div>

          <div className="day-tabs no-print" role="group" aria-label="Filter by day">
            <button
              type="button"
              className="day-tab"
              aria-pressed={activeDay === null}
              onClick={() => setActiveDay(null)}
            >
              All days
            </button>
            {plan.map((day, i) => (
              <button
                key={i}
                type="button"
                className="day-tab"
                aria-pressed={activeDay === i + 1}
                onClick={() => setActiveDay(i + 1)}
              >
                Day {i + 1}
                <span className="muted small mono"> · {day.length}</span>
              </button>
            ))}
          </div>

          <div className="results-layout">
            <div>
              {visibleDays.map(({ dayNumber, stops }) => (
                <ItineraryDay
                  key={dayNumber}
                  day={stops}
                  dayNumber={dayNumber}
                  activeId={activeId}
                  onSelect={select}
                />
              ))}
            </div>

            <div className="map-shell no-print">
              <Map
                points={mapPoints}
                activeId={activeId}
                onSelect={setActiveId}
                showRoute={activeDay !== null}
                height={520}
              />
            </div>
          </div>
        </>
      )}
    </div>
  )
}
