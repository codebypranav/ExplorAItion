"use client"

import { useRouter } from 'next/navigation'
import React, { useState } from 'react'

const EXAMPLES = [
  '3-day trip to Paris with art museums and good coffee',
  'family-friendly places in Tokyo with scenic views',
  'historic neighbourhoods and food spots in Rome',
  'quiet hikes and lakes near Zurich',
]

export default function HeroSearch() {
  const router = useRouter()
  const [query, setQuery] = useState('')

  const go = (q: string) => {
    const trimmed = q.trim()
    if (!trimmed) return
    router.push(`/search?q=${encodeURIComponent(trimmed)}`)
  }

  return (
    <div className="stack">
      <form
        className="panel"
        onSubmit={(e) => {
          e.preventDefault()
          go(query)
        }}
      >
        <div className="field">
          <label htmlFor="hero-query">Describe the trip you want</label>
          <textarea
            id="hero-query"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={(e) => {
              // Enter submits; Shift+Enter keeps the newline.
              if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault()
                go(query)
              }
            }}
            placeholder="e.g. four days in Lisbon — tiled façades, viewpoints, and seafood near the water"
            rows={3}
          />
          <span className="hint">
            Plain language works best. Press Enter to search.
          </span>
        </div>
        <div className="row-between" style={{ marginTop: '1rem' }}>
          <span className="small muted">
            Semantic search, not keyword matching.
          </span>
          <button type="submit" className="btn btn-primary" disabled={!query.trim()}>
            Find places
            <span aria-hidden="true">→</span>
          </button>
        </div>
      </form>

      <div className="row" style={{ justifyContent: 'center' }}>
        <span className="eyebrow">Try</span>
        {EXAMPLES.map((example) => (
          <button
            key={example}
            type="button"
            className="chip"
            onClick={() => {
              setQuery(example)
              go(example)
            }}
          >
            {example}
          </button>
        ))}
      </div>
    </div>
  )
}
