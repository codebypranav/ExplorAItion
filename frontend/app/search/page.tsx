"use client"
import React, { useState } from 'react'
import dynamic from 'next/dynamic'
import ResultCard from '../../components/ResultCard'
const Map = dynamic(() => import('../../components/Map'), { ssr: false })

export default function SearchPage() {
  const apiBase = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
  const [query, setQuery] = useState('')
  const [topk, setTopk] = useState(10)
  const [results, setResults] = useState([])
  const [mapCenter, setMapCenter] = useState(null)
  const [loading, setLoading] = useState(false)

  async function doSearch(e) {
    e.preventDefault()
    setLoading(true)
    // Send query as is, let backend parse it
    const body = { query, filters: {}, top_k: topk }
    try {
      const res = await fetch(`${apiBase}/recommend`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) })
      if (!res.ok) {
        throw new Error(`recommend failed with status ${res.status}`)
      }
      const json = await res.json()
      setResults(Array.isArray(json) ? json : [])
      if (json.length>0 && json[0].latitude && json[0].longitude) {
        setMapCenter({ latitude: json[0].latitude, longitude: json[0].longitude })
      }
    } catch (err) {
      console.error(err)
      setResults([])
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ padding: '2rem', maxWidth: 1200, margin: '0 auto' }}>
      <div style={{ textAlign: 'center', marginBottom: '3rem' }}>
        <h2 style={{ fontSize: '2.5rem', marginBottom: '1rem', color: '#2E4600' }}>Find Your Next Adventure</h2>
        <p style={{ fontSize: '1.2rem', fontStyle: 'italic', color: '#5D4037' }}>Tell us what you're dreaming of...</p>
      </div>
      
      <form onSubmit={doSearch} style={{ maxWidth: 800, margin: '0 auto', background: '#fff', padding: '2rem', borderRadius: '8px', boxShadow: '0 4px 12px rgba(0,0,0,0.1)', border: '1px solid #8D6E63' }}>
        <div style={{ display: 'flex', gap: '1rem', flexDirection: 'column' }}>
          <textarea 
            value={query} 
            onChange={(e)=>setQuery(e.target.value)} 
            placeholder="e.g. I want to visit art museums in Paris and drink great coffee, or maybe go hiking in the Swiss Alps." 
            style={{ width: '100%', padding: '16px', fontSize: '1.1rem', minHeight: '100px', fontFamily: 'inherit', border: '2px solid #D7CCC8', borderRadius: '6px' }} 
          />
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <label style={{ color: '#5D4037' }}>Results:</label>
              <input type='number' value={topk} onChange={(e)=>setTopk(Number(e.target.value))} style={{ width: 60, padding: '8px' }} />
            </div>
            <button type='submit' disabled={loading} style={{ fontSize: '1.1rem', padding: '12px 32px', opacity: loading ? 0.7 : 1 }}>
              {loading ? 'Exploring...' : 'Search'}
            </button>
          </div>
        </div>
      </form>

      <div style={{ marginTop: '3rem', display: 'flex', gap: 24, flexDirection: 'column-reverse' }}>
        {results && results.length>0 && (
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))', gap: '24px' }}>
            {results.map((r, i)=> (
              <div key={r.xid || i} style={{ background: '#fff', borderRadius: '8px', overflow: 'hidden', boxShadow: '0 2px 8px rgba(0,0,0,0.05)', border: '1px solid #EFEBE9' }}>
                <ResultCard place={r} onCenter={(p) => setMapCenter({ latitude: p.latitude, longitude: p.longitude })} />
              </div>
            ))}
          </div>
        )}
        
        <div style={{ height: '400px', borderRadius: '12px', overflow: 'hidden', border: '4px solid #fff', boxShadow: '0 4px 12px rgba(0,0,0,0.1)' }}>
           <Map points={results.map(r => ({ latitude: r.latitude, longitude: r.longitude, name: r.name, rating: r.rating, description: r.description }))} center={mapCenter ? [mapCenter.latitude, mapCenter.longitude] : null} />
        </div>
      </div>
    </div>
  )
}
