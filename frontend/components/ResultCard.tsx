import React from 'react'

export default function ResultCard({ place, onCenter }) {
  return (
    <div style={{ display: 'flex', gap: '12px', alignItems: 'center' }}>
      {place.image_url ? <img src={place.image_url} style={{ width: 140, height: 90, objectFit: 'cover' }} alt=''/> : <div style={{ width: 140, height: 90, background: '#f2f2f2' }} />}
      <div>
        <div style={{ fontWeight: 700 }}>{place.name}</div>
        <div style={{ opacity: 0.8 }}>{place.description}</div>
        <div style={{ opacity: 0.6 }}>{place.country} • Score {Number(place.score).toFixed(3)} {place.rating ? ` • Rating ${place.rating}` : ''}</div>
        <div style={{ marginTop: 8 }}>
          <button onClick={() => onCenter && onCenter(place)} style={{ padding: '6px 8px' }}>Center on map</button>
        </div>
      </div>
    </div>
  )
}
