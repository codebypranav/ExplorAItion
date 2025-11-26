"use client"
import React, { useEffect, useState } from 'react'

export default function ItineraryPage(){
  const [city,setCity] = useState('Paris')
  const [days,setDays] = useState(2)
  const [query,setQuery] = useState('')
  const [itinerary,setItinerary] = useState([])

  async function fetchItinerary(e){
    e.preventDefault()
    const body = { city, days, query }
    const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/itinerary`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) })
    const json = await res.json()
    setItinerary(json)
  }

  return (
    <div style={{ padding: '2rem' }}>
      <h2>Generate Itinerary</h2>
      <form onSubmit={fetchItinerary} style={{ maxWidth: 600 }}>
        <div style={{ marginBottom: 8 }}>
          <input value={city} onChange={(e)=>setCity(e.target.value)} placeholder='City' />
        </div>
        <div style={{ marginBottom: 8 }}>
          <input value={days} type='number' onChange={(e)=>setDays(Number(e.target.value))} placeholder='Days' />
        </div>
        <div style={{ marginBottom: 8 }}>
          <input value={query} placeholder='Interests' onChange={(e)=>setQuery(e.target.value)} />
        </div>
        <button type='submit'>Generate</button>
      </form>
      <div style={{ marginTop: 24 }}>
        {itinerary && itinerary.length>0 && itinerary.map((day, idx)=> (
          <div key={idx} style={{ marginBottom: 18 }}>
            <div style={{ fontWeight: 600 }}>Day {idx+1}</div>
            <ul>
              {day.map((p)=> (
                <li key={p.xid} style={{ padding: '6px 0' }}>
                  {p.name} — {p.latitude}, {p.longitude}
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>
    </div>
  )
}
