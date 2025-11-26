"use client"
import React, { useState } from 'react'
import useSWR from 'swr'
import dynamic from 'next/dynamic'
import ResultCard from '../../components/ResultCard'
const Map = dynamic(() => import('../../components/Map'), { ssr: false })

const fetcher = (url, opts) => fetch(url, opts).then(r => r.json())

export default function SearchPage() {
  const [query, setQuery] = useState('')
  const [filters, setFilters] = useState('{}')
  const [topk, setTopk] = useState(10)
  const [results, setResults] = useState([])
  const [mapCenter, setMapCenter] = useState(null)

  async function doSearch(e) {
    e.preventDefault()
    const body = { query, filters: JSON.parse(filters || '{}'), top_k: topk }
    const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/recommend`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) })
    const json = await res.json()
    setResults(json)
    // set the map center to first result (if exists)
    if (json.length>0 && json[0].latitude && json[0].longitude) {
      setMapCenter({ latitude: json[0].latitude, longitude: json[0].longitude })
    }
  }

  return (
    <div style={{ padding: '1rem' }}>
      <h2>Search</h2>
      <form onSubmit={doSearch} style={{ maxWidth: 800 }}>
        <input type="text" value={query} onChange={(e)=>setQuery(e.target.value)} placeholder="I like hiking and coffee" style={{ width: '100%', padding: '12px' }} />
        <div style={{ marginTop: '8px' }}>
          <input type='text' value={filters} onChange={(e)=>setFilters(e.target.value)} placeholder='{"country":"France"}' style={{ width:'70%', padding: '8px' }} />
          <input type='number' value={topk} onChange={(e)=>setTopk(Number(e.target.value))} style={{ marginLeft: '8px', width: 80 }} />
          <button type='submit' style={{ marginLeft: '8px' }}>Search</button>
        </div>
      </form>
      <div style={{ marginTop: '24px', display: 'flex', gap: 12 }}>
        {results && results.length>0 && <ul style={{ listStyle: 'none', padding: 0 }}>
          {results.map((r, i)=> (
            <li key={r.xid || i} style={{ padding: '12px', borderBottom: '1px solid #eee' }}>
              <ResultCard place={r} onCenter={(p) => setMapCenter({ latitude: p.latitude, longitude: p.longitude })} />
            </li>
          ))}
        </ul> }
        <div style={{ flex: 1 }}>
          <Map points={results.map(r => ({ latitude: r.latitude, longitude: r.longitude, name: r.name, rating: r.rating, description: r.description }))} center={mapCenter ? [mapCenter.latitude, mapCenter.longitude] : null} />
        </div>
      </div>
    </div>
  )
}
