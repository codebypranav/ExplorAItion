import Link from 'next/link'
import React from 'react'

import HeroSearch from '../components/HeroSearch'

const PIPELINE = [
  {
    step: 'Step 1',
    title: 'Read the request',
    body: 'Your sentence is parsed into a structured query — place, interests, constraints.',
  },
  {
    step: 'Step 2',
    title: 'Search by meaning',
    body: 'The query becomes a sparse embedding and is matched against the POI index in Pinecone.',
  },
  {
    step: 'Step 3',
    title: 'Rank and enrich',
    body: 'Matches are blended with Google Places ratings and current Open-Meteo weather.',
  },
  {
    step: 'Step 4',
    title: 'Route the days',
    body: 'A nearest-neighbour pass orders stops so each day stays geographically sane.',
  },
]

export default function Home() {
  return (
    <div className="shell page">
      <section className="hero">
        <p className="eyebrow">AI trip planning</p>
        <h1>Describe the trip. Get the route.</h1>
        <p className="lede">
          ExplorAItion turns a sentence into matching places and a day-by-day
          plan — semantic search over a point-of-interest index, then a route
          that keeps nearby stops together.
        </p>
      </section>

      <HeroSearch />

      <section className="stack-lg" style={{ marginTop: '4rem' }}>
        <div>
          <p className="eyebrow">How it works</p>
          <h2>Four stages, one sentence of input</h2>
        </div>
        <div className="feature-grid">
          {PIPELINE.map((item) => (
            <article className="feature" key={item.step}>
              <span className="step">{item.step}</span>
              <h3>{item.title}</h3>
              <p>{item.body}</p>
            </article>
          ))}
        </div>
      </section>

      <section
        className="feature-grid"
        style={{ marginTop: '2.5rem', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))' }}
      >
        <article className="panel stack">
          <div>
            <p className="eyebrow">Explore</p>
            <h3 style={{ fontSize: '1.3rem' }}>Search places</h3>
            <p style={{ margin: 0 }}>
              Rank individual POIs by how well they match an interest, with
              ratings, live weather and a map you can pan from the results list.
            </p>
          </div>
          <Link href="/search" className="btn btn-primary" style={{ justifySelf: 'start' }}>
            Open search
          </Link>
        </article>

        <article className="panel stack">
          <div>
            <p className="eyebrow">Plan</p>
            <h3 style={{ fontSize: '1.3rem' }}>Build an itinerary</h3>
            <p style={{ margin: 0 }}>
              Pick a city, a number of days and your interests, and get a routed
              plan with per-day distances and a drawn route.
            </p>
          </div>
          <Link href="/itinerary" className="btn" style={{ justifySelf: 'start' }}>
            Open itinerary planner
          </Link>
        </article>
      </section>
    </div>
  )
}
