import React from 'react'
import Link from 'next/link'

export default function Home() {
  return (
    <main style={{ padding: '2rem' }}>
      <h1>ExplorAItion</h1>
      <p>Welcome to ExplorAItion — demo frontend</p>
      <div style={{ marginTop: '1rem' }}>
        <Link href="/search">Search</Link>
      </div>
      <div style={{ marginTop: '0.5rem' }}>
        <Link href="/itinerary">Itinerary</Link>
      </div>
    </main>
  )
}
