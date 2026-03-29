"use client"
import React, { useState } from 'react'
import dynamic from 'next/dynamic'
import ResultCard from '../../components/ResultCard'

const Map = dynamic(() => import('../../components/Map'), { ssr: false })

export default function ItineraryPage(){
  const apiBase = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
  const [city,setCity] = useState('Paris')
  const [days,setDays] = useState(2)
  const [query,setQuery] = useState('')
  const [itinerary,setItinerary] = useState([])
  const [mapCenter, setMapCenter] = useState(null)

  async function fetchItinerary(e){
    e.preventDefault()
    const body = { city, days, query }
    try {
      const res = await fetch(`${apiBase}/itinerary`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) })
      if (!res.ok) {
        throw new Error(`itinerary failed with status ${res.status}`)
      }
      const json = await res.json()
      const safeItinerary = Array.isArray(json) ? json : []
      setItinerary(safeItinerary)
      // Center map on the first place of the first day
      if (safeItinerary.length > 0 && safeItinerary[0].length > 0) {
        setMapCenter({ latitude: safeItinerary[0][0].latitude, longitude: safeItinerary[0][0].longitude })
      }
    } catch (err) {
      console.error(err)
      setItinerary([])
    }
  }

  // Flatten itinerary for map points
  const allPoints = itinerary.flat().map(p => ({
    latitude: p.latitude,
    longitude: p.longitude,
    name: p.name,
    rating: p.rating,
    description: p.description
  }))

  return (
    <div style={{ padding: '2rem' }}>
      <h2>Generate Itinerary</h2>
      <form onSubmit={fetchItinerary} style={{ maxWidth: 600 }}>
        <div style={{ marginBottom: 8 }}>
          <input value={city} onChange={(e)=>setCity(e.target.value)} placeholder='City' style={{ padding: '8px', width: '100%' }} />
        </div>
        <div style={{ marginBottom: 8 }}>
          <input value={days} type='number' onChange={(e)=>setDays(Number(e.target.value))} placeholder='Days' style={{ padding: '8px', width: '100%' }} />
        </div>
        <div style={{ marginBottom: 8 }}>
          <input value={query} placeholder='Interests (e.g. museums, food)' onChange={(e)=>setQuery(e.target.value)} style={{ padding: '8px', width: '100%' }} />
        </div>
        <button type='submit' style={{ padding: '8px 16px' }}>Generate</button>
      </form>
      
      <div style={{ marginTop: 24, display: 'flex', gap: 24 }}>
        <div style={{ flex: 1 }}>
          {itinerary && itinerary.length>0 && itinerary.map((day, idx)=> (
            <div key={idx} style={{ marginBottom: 24 }}>
              <h3 style={{ fontWeight: 600, marginBottom: 12 }}>Day {idx+1}</h3>
              <ul style={{ listStyle: 'none', padding: 0 }}>
                {day.map((p)=> (
                  <li key={p.xid} style={{ padding: '12px', borderBottom: '1px solid #eee' }}>
                    <ResultCard place={p} onCenter={(p) => setMapCenter({ latitude: p.latitude, longitude: p.longitude })} />
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
        <div style={{ flex: 1, position: 'sticky', top: 20, height: 'fit-content' }}>
           {itinerary.length > 0 && (
             <Map 
               points={allPoints} 
               center={mapCenter ? [mapCenter.latitude, mapCenter.longitude] : null} 
             />
           )}
        </div>
      </div>
    </div>
  )
}
