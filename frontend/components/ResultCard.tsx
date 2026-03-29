import React from 'react'

export default function ResultCard({ place, onCenter }) {
  return (
    <div style={{ display: 'flex', gap: '16px', alignItems: 'start' }}>
      {place.image_url ? (
        <img 
          src={place.image_url} 
          style={{ width: 160, height: 120, objectFit: 'cover', borderRadius: '4px', border: '1px solid var(--accent-wood)' }} 
          alt={place.name}
        />
      ) : (
        <div style={{ width: 160, height: 120, background: 'rgba(141, 110, 99, 0.2)', borderRadius: '4px', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--accent-wood)' }}>
          No Image
        </div>
      )}
      <div style={{ flex: 1 }}>
        <div style={{ fontWeight: 700, fontSize: '1.2rem', color: 'var(--primary-green)', fontFamily: 'Courier New, monospace' }}>{place.name}</div>
        <div style={{ margin: '4px 0', fontStyle: 'italic', color: '#555' }}>{place.country}</div>
        <div style={{ opacity: 0.9, lineHeight: '1.4' }}>{place.description}</div>
        <div style={{ marginTop: '8px', fontSize: '0.9rem', color: 'var(--accent-wood)' }}>
           Score: {Number(place.score).toFixed(3)} {place.rating ? ` • Rating: ${place.rating}/5` : ''}
        </div>
        <div style={{ marginTop: 12 }}>
          <button onClick={() => onCenter && onCenter(place)} style={{ padding: '6px 12px', fontSize: '0.8rem' }}>Locate on Map</button>
        </div>
      </div>
    </div>
  )
}
