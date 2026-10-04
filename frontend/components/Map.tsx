"use client"
import React, { useEffect } from 'react'
import { MapContainer, TileLayer, Marker, Popup, useMap } from 'react-leaflet'
import '../lib/leaflet' // configure default icons

export type MapPoint = {
  latitude: number
  longitude: number
  name?: string
  rating?: number | string
  description?: string
}

type LatLng = [number, number]

type MapProps = {
  points?: MapPoint[]
  center?: LatLng | null
  onMarkerClick?: ((point: MapPoint) => void) | null
}

// react-leaflet v5 has no `whenCreated`; the map instance comes from useMap().
function Recenter({ center }: { center: LatLng | null }) {
  const map = useMap()
  const lat = center?.[0]
  const lng = center?.[1]
  useEffect(() => {
    if (lat !== undefined && lng !== undefined) map.setView([lat, lng], 13)
  }, [map, lat, lng])
  return null
}

function MarkerFocus({ point, onMarkerClick }: { point: MapPoint; onMarkerClick: MapProps['onMarkerClick'] }) {
  const map = useMap()
  return (
    <Marker
      position={[point.latitude, point.longitude]}
      eventHandlers={{
        click: () => {
          map.setView([point.latitude, point.longitude], 13)
          if (onMarkerClick) onMarkerClick(point)
        },
      }}
    >
      <Popup>
        <div style={{ maxWidth: 220 }}>
          <div style={{ fontWeight: 700 }}>{point.name}</div>
          {point.rating && <div>Rating: {point.rating}</div>}
          <div>{point.description}</div>
        </div>
      </Popup>
    </Marker>
  )
}

export default function Map({ points = [], center = null, onMarkerClick = null }: MapProps) {
  if (!points || points.length === 0) return <div>No points to show</div>
  const initialCenter: LatLng = center || [points[0].latitude, points[0].longitude]

  return (
    <div style={{ height: 400, width: '100%' }}>
      <MapContainer center={initialCenter} zoom={13} style={{ height: 400, width: '100%' }}>
        <TileLayer url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png" />
        <Recenter center={center} />
        {points.map((p, i) => (
          <MarkerFocus key={i} point={p} onMarkerClick={onMarkerClick} />
        ))}
      </MapContainer>
    </div>
  )
}
