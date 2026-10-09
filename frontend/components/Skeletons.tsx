import React from 'react'

export function PlaceSkeleton() {
  return (
    <li className="place-card" aria-hidden="true">
      <div className="skeleton" style={{ aspectRatio: '4 / 3' }} />
      <div style={{ display: 'grid', gap: 8, alignContent: 'start' }}>
        <div className="skeleton" style={{ height: 18, width: '55%' }} />
        <div className="skeleton" style={{ height: 12, width: '92%' }} />
        <div className="skeleton" style={{ height: 12, width: '80%' }} />
        <div className="skeleton" style={{ height: 22, width: 160, borderRadius: 999 }} />
      </div>
    </li>
  )
}

export function PlaceListSkeleton({ count = 4 }: { count?: number }) {
  return (
    <ul className="place-list" aria-busy="true" aria-label="Loading results">
      {Array.from({ length: count }, (_, i) => (
        <PlaceSkeleton key={i} />
      ))}
    </ul>
  )
}
