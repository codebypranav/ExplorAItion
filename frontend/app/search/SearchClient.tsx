"use client"

import dynamic from 'next/dynamic'
import { useRouter, useSearchParams } from 'next/navigation'
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import ResultCard from '../../components/ResultCard'
import EmptyState from '../../components/EmptyState'
import { DemoNotice, ErrorNotice } from '../../components/Notices'
import { PlaceListSkeleton } from '../../components/Skeletons'
import { LIMITS, recommend } from '../../lib/api'
import { samplePlaces, SAMPLE_QUERY } from '../../lib/sample'
import type { Place } from '../../lib/types'

const Map = dynamic(() => import('../../components/Map'), {
  ssr: false,
  loading: () => <div className="skeleton" style={{ height: 440 }} />,
})

type SortKey = 'match' | 'rating' | 'name'

const DEMO_DEFAULT = process.env.NEXT_PUBLIC_DEMO_MODE === 'true'

export default function SearchClient() {
  const router = useRouter()
  const params = useSearchParams()

  const urlQuery = params.get('q') ?? ''
  const urlTopK = Number(params.get('k')) || LIMITS.topK.default
  const urlCountry = params.get('country') ?? ''
  const urlMinRating = params.get('min_rating') ?? ''

  // Form state is local; the URL is the source of truth for what was searched.
  const [query, setQuery] = useState(urlQuery)
  const [topK, setTopK] = useState(urlTopK)
  const [country, setCountry] = useState(urlCountry)
  const [minRating, setMinRating] = useState(urlMinRating)
  const [showFilters, setShowFilters] = useState(!!urlCountry || !!urlMinRating)

  const [results, setResults] = useState<Place[] | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<unknown>(null)
  const [isSample, setIsSample] = useState(false)
  const [sort, setSort] = useState<SortKey>('match')
  const [activeId, setActiveId] = useState<string | null>(null)

  const inFlight = useRef<AbortController | null>(null)

  const loadSample = useCallback(() => {
    inFlight.current?.abort()
    setLoading(false)
    setError(null)
    setIsSample(true)
    setResults(samplePlaces(topK))
    setActiveId(null)
  }, [topK])

  const runSearch = useCallback(
    async (q: string, k: number, filters: Record<string, unknown>) => {
      inFlight.current?.abort()
      const controller = new AbortController()
      inFlight.current = controller

      setLoading(true)
      setError(null)
      setIsSample(false)
      setActiveId(null)

      try {
        const places = await recommend({ query: q, top_k: k, filters }, controller.signal)
        setResults(places)
      } catch (err) {
        if (controller.signal.aborted) return
        setError(err)
        setResults(null)
      } finally {
        if (!controller.signal.aborted) setLoading(false)
      }
    },
    [],
  )

  // Build the Pinecone metadata filter from the simple UI controls.
  const filters = useMemo(() => {
    const f: Record<string, unknown> = {}
    const c = urlCountry.trim().toUpperCase()
    if (c) f.country = { $eq: c }
    const r = Number(urlMinRating)
    if (Number.isFinite(r) && r > 0) f.rate = { $gte: r }
    return f
  }, [urlCountry, urlMinRating])

  // Search whenever the URL says to — covers deep links and the back button.
  useEffect(() => {
    if (!urlQuery.trim()) {
      setResults(null)
      return
    }
    if (DEMO_DEFAULT) {
      setIsSample(true)
      setResults(samplePlaces(urlTopK))
      return
    }
    runSearch(urlQuery, urlTopK, filters)
  }, [urlQuery, urlTopK, filters, runSearch])

  useEffect(() => () => inFlight.current?.abort(), [])

  const submit = () => {
    const trimmed = query.trim()
    if (!trimmed) return
    const next = new URLSearchParams({ q: trimmed, k: String(topK) })
    if (country.trim()) next.set('country', country.trim().toUpperCase())
    if (Number(minRating) > 0) next.set('min_rating', minRating)

    if (next.toString() === params.toString()) {
      // Same URL, so the effect below will not re-run — search directly.
      runSearch(trimmed, topK, filters)
      return
    }
    router.push(`/search?${next.toString()}`)
  }

  const sorted = useMemo(() => {
    if (!results) return null
    const copy = [...results]
    if (sort === 'rating') {
      copy.sort((a, b) => (b.rating ?? 0) - (a.rating ?? 0) || b.score - a.score)
    } else if (sort === 'name') {
      copy.sort((a, b) => (a.name || '').localeCompare(b.name || ''))
    }
    // 'match' keeps the backend's own composite ordering.
    return copy
  }, [results, sort])

  const mapPoints = useMemo(
    () =>
      (sorted ?? []).map((p, i) => ({
        id: p.xid || `${p.name}-${i}`,
        name: p.name,
        latitude: p.latitude,
        longitude: p.longitude,
        description: p.description,
        rating: p.rating,
        weather: p.weather,
        label: i + 1,
      })),
    [sorted],
  )

  const idOf = (p: Place, i: number) => p.xid || `${p.name}-${i}`

  return (
    <div className="shell page stack-lg">
      <header>
        <p className="eyebrow">Semantic search</p>
        <h1 style={{ marginBottom: '0.3rem' }}>Find places worth the detour</h1>
        <p style={{ maxWidth: '62ch' }}>
          Describe what you want to see. Results are ranked by meaning, then
          blended with ratings — not matched on keywords.
        </p>
      </header>

      <form
        className="panel stack"
        onSubmit={(e) => {
          e.preventDefault()
          submit()
        }}
      >
        <div className="field">
          <label htmlFor="query">Your request</label>
          <textarea
            id="query"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault()
                submit()
              }
            }}
            placeholder={`e.g. ${SAMPLE_QUERY}`}
            rows={3}
          />
        </div>

        <div className="row-between">
          <div className="row">
            <div className="field" style={{ width: 150 }}>
              <label htmlFor="topk">
                Results: <span className="mono">{topK}</span>
              </label>
              <input
                id="topk"
                type="range"
                min={LIMITS.topK.min}
                max={LIMITS.topK.max}
                value={topK}
                onChange={(e) => setTopK(Number(e.target.value))}
              />
            </div>
            <button
              type="button"
              className="btn btn-ghost btn-sm"
              aria-expanded={showFilters}
              onClick={() => setShowFilters((v) => !v)}
            >
              {showFilters ? 'Hide filters' : 'Filters'}
              <span aria-hidden="true">{showFilters ? '▲' : '▼'}</span>
            </button>
          </div>

          <button type="submit" className="btn btn-primary" disabled={loading || !query.trim()}>
            {loading ? 'Searching…' : 'Search'}
          </button>
        </div>

        {showFilters && (
          <div
            className="row"
            style={{ alignItems: 'flex-end', gap: '1rem', borderTop: '1px solid var(--line)', paddingTop: '1rem' }}
          >
            <div className="field" style={{ width: 170 }}>
              <label htmlFor="country">Country code</label>
              <input
                id="country"
                value={country}
                onChange={(e) => setCountry(e.target.value)}
                placeholder="FR"
                maxLength={2}
              />
              <span className="hint">Metadata filter on `country`</span>
            </div>
            <div className="field" style={{ width: 170 }}>
              <label htmlFor="minrating">Minimum rate</label>
              <input
                id="minrating"
                type="number"
                min={0}
                max={7}
                step={1}
                value={minRating}
                onChange={(e) => setMinRating(e.target.value)}
                placeholder="2"
              />
              <span className="hint">OpenTripMap `rate` ≥ value</span>
            </div>
            {(country || minRating) && (
              <button
                type="button"
                className="btn btn-ghost btn-sm"
                onClick={() => {
                  setCountry('')
                  setMinRating('')
                }}
              >
                Clear
              </button>
            )}
          </div>
        )}
      </form>

      {isSample && <DemoNotice onExit={() => runSearch(urlQuery || SAMPLE_QUERY, topK, filters)} />}
      {error ? (
        <ErrorNotice
          error={error}
          onRetry={() => runSearch(urlQuery, urlTopK, filters)}
          onUseSample={loadSample}
        />
      ) : null}

      {loading && <PlaceListSkeleton count={Math.min(4, topK)} />}

      {!loading && !error && sorted === null && (
        <EmptyState emoji="🔭" title="Nothing searched yet">
          Enter a request above, or{' '}
          <button
            type="button"
            className="btn btn-sm"
            onClick={() => {
              setQuery(SAMPLE_QUERY)
              router.push(`/search?q=${encodeURIComponent(SAMPLE_QUERY)}&k=${topK}`)
            }}
          >
            try an example
          </button>
        </EmptyState>
      )}

      {!loading && !error && sorted?.length === 0 && (
        <EmptyState emoji="🫙" title="No matches">
          The index returned nothing for this query. Loosen the filters, or check
          that the Pinecone index has been seeded.
        </EmptyState>
      )}

      {!loading && sorted && sorted.length > 0 && (
        <>
          <div className="row-between">
            <span className="small muted">
              <strong className="mono">{sorted.length}</strong> place
              {sorted.length === 1 ? '' : 's'} for “{urlQuery || SAMPLE_QUERY}”
            </span>
            <div className="row no-print">
              <label htmlFor="sort" className="small muted">
                Sort
              </label>
              <select
                id="sort"
                value={sort}
                onChange={(e) => setSort(e.target.value as SortKey)}
                style={{ width: 'auto' }}
              >
                <option value="match">Best match</option>
                <option value="rating">Highest rated</option>
                <option value="name">Name (A–Z)</option>
              </select>
            </div>
          </div>

          <div className="results-layout">
            <ul className="place-list">
              {sorted.map((place, i) => (
                <ResultCard
                  key={idOf(place, i)}
                  place={place}
                  active={activeId === idOf(place, i)}
                  onSelect={() => setActiveId(idOf(place, i))}
                />
              ))}
            </ul>

            <div className="map-shell no-print">
              <Map
                points={mapPoints}
                activeId={activeId}
                onSelect={setActiveId}
                height={480}
              />
            </div>
          </div>
        </>
      )}
    </div>
  )
}
