"use client"
import React, { useEffect, useRef } from 'react'
import dynamic from 'next/dynamic'
import '../lib/leaflet' // configure default icons

const MapContainer = dynamic(() => import('react-leaflet').then(mod => mod.MapContainer), { ssr: false })
const TileLayer = dynamic(() => import('react-leaflet').then(mod => mod.TileLayer), { ssr: false })
const Marker = dynamic(() => import('react-leaflet').then(mod => mod.Marker), { ssr: false })
const Popup = dynamic(() => import('react-leaflet').then(mod => mod.Popup), { ssr: false })

export default function Map({ points = [], center = null, onMarkerClick = null }){
  // points: [{ latitude, longitude, name }]
  if (!points || points.length === 0) return <div>No points to show</div>
  const inferredCenter = center || [points[0].latitude, points[0].longitude]
  const mapRef = useRef<any>(null)

  useEffect(() => {
    const m = mapRef.current
    if (!m) return
    if (center && Array.isArray(center) && center.length === 2) {
      m.setView(center, 13)
    }
  }, [center])

  const MapContainerAny: any = MapContainer
  return (
    <div style={{ height: 400, width: '100%' }}>
      <MapContainerAny zoom={13} style={{ height: 400, width: '100%' }} whenCreated={(m:any) => { mapRef.current = m; if (inferredCenter && Array.isArray(inferredCenter) && inferredCenter.length === 2) m.setView(inferredCenter, 13) }}>
        <TileLayer url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png" />
        {points.map((p, i) => (
          <Marker key={i} position={[p.latitude, p.longitude]} eventHandlers={{ click: () => {
            if (mapRef.current) {
              mapRef.current.setView([p.latitude, p.longitude], 13)
            }
            if (onMarkerClick) onMarkerClick(p)
          } }}>
            <Popup>
              <div style={{ maxWidth: 220 }}>
                <div style={{ fontWeight: 700 }}>{p.name}</div>
                {p.rating && <div>Rating: {p.rating}</div>}
                <div>{p.description}</div>
              </div>
            </Popup>
          </Marker>
        ))}
      </MapContainerAny>
    </div>
  )
}
