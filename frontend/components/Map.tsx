"use client"

import 'leaflet/dist/leaflet.css'

import L from 'leaflet'
import React, { useEffect, useMemo, useState } from 'react'
import {
  MapContainer,
  Marker,
  Polyline,
  Popup,
  TileLayer,
  useMap,
} from 'react-leaflet'

import { hasCoords } from '../lib/api'
import { bounds as boundsOf } from '../lib/geo'
import { describeWeather } from '../lib/weather'
import type { Weather } from '../lib/types'

export type MapPoint = {
  id: string
  name: string
  latitude: number
  longitude: number
  description?: string
  rating?: number
  weather?: Weather
  /** Shown inside the pin; falls back to a dot when omitted. */
  label?: number
}

type MapProps = {
  points: MapPoint[]
  /** Pin id to highlight and pan to. */
  activeId?: string | null
  onSelect?: (id: string) => void
  /** Draws a dashed route through the points in the order given. */
  showRoute?: boolean
  height?: number | string
}

const TILES = {
  url: 'https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png',
  attribution:
    '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
}

function pinIcon(label: number | undefined, active: boolean) {
  return L.divIcon({
    className: 'pin-wrap',
    html: `<div class="pin" data-active="${active}"><span>${label ?? '•'}</span></div>`,
    iconSize: active ? [34, 34] : [28, 28],
    iconAnchor: active ? [17, 34] : [14, 28],
    popupAnchor: [0, -26],
  })
}

/** Keeps the viewport in sync with the points and the active selection. */
function Viewport({
  points,
  activeId,
}: {
  points: MapPoint[]
  activeId?: string | null
}) {
  const map = useMap()
  const key = points.map((p) => p.id).join('|')

  useEffect(() => {
    const box = boundsOf(points)
    if (!box) return
    const [[s, w], [n, e]] = box
    if (s === n && w === e) {
      map.setView([s, w], 14)
    } else {
      map.fitBounds(box, { padding: [40, 40], maxZoom: 15 })
    }
    // Leaflet needs a nudge when the container was hidden or resized.
    map.invalidateSize()
  }, [key, map]) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (!activeId) return
    const target = points.find((p) => p.id === activeId)
    if (!target) return
    map.flyTo([target.latitude, target.longitude], Math.max(map.getZoom(), 14), {
      duration: 0.6,
    })
  }, [activeId, key, map]) // eslint-disable-line react-hooks/exhaustive-deps

  return null
}

export default function Map({
  points,
  activeId = null,
  onSelect,
  showRoute = false,
  height = 440,
}: MapProps) {
  const plotted = useMemo(() => points.filter(hasCoords), [points])
  const [theme, setTheme] = useState<'light' | 'dark'>('light')

  useEffect(() => {
    const mq = window.matchMedia('(prefers-color-scheme: dark)')
    const sync = () => setTheme(mq.matches ? 'dark' : 'light')
    sync()
    mq.addEventListener('change', sync)
    return () => mq.removeEventListener('change', sync)
  }, [])

  if (plotted.length === 0) {
    return (
      <div
        style={{ height, display: 'grid', placeItems: 'center', padding: '1.5rem' }}
        className="muted small"
      >
        No mappable coordinates yet.
      </div>
    )
  }

  const route: Array<[number, number]> = plotted.map((p) => [p.latitude, p.longitude])

  return (
    <MapContainer
      center={[plotted[0].latitude, plotted[0].longitude]}
      zoom={13}
      scrollWheelZoom
      style={{ height, width: '100%' }}
    >
      {/* OSM raster tiles are tinted for dark mode in globals.css — free dark
          basemaps now need an API key, so filtering the light tiles is simpler. */}
      <TileLayer url={TILES.url} attribution={TILES.attribution} maxZoom={19} />
      {showRoute && route.length > 1 && (
        <Polyline
          positions={route}
          pathOptions={{
            color: theme === 'dark' ? '#7fb98c' : '#2f5d3a',
            weight: 3,
            opacity: 0.75,
            dashArray: '6 8',
          }}
        />
      )}
      {plotted.map((p) => (
        <Marker
          key={p.id}
          position={[p.latitude, p.longitude]}
          icon={pinIcon(p.label, p.id === activeId)}
          zIndexOffset={p.id === activeId ? 1000 : 0}
          eventHandlers={{ click: () => onSelect?.(p.id) }}
        >
          <Popup>
            <div style={{ maxWidth: 240 }}>
              <b>{p.name || 'Unnamed place'}</b>
              {p.description && (
                <div style={{ marginTop: 4 }}>{truncate(p.description, 160)}</div>
              )}
              <div style={{ marginTop: 6, opacity: 0.75, fontSize: '0.8rem' }}>
                {p.rating ? `★ ${p.rating.toFixed(1)}` : null}
                {p.rating && p.weather ? ' · ' : null}
                {p.weather
                  ? `${describeWeather(p.weather).icon} ${describeWeather(p.weather).celsius}°C`
                  : null}
              </div>
            </div>
          </Popup>
        </Marker>
      ))}
      <Viewport points={plotted} activeId={activeId} />
    </MapContainer>
  )
}

function truncate(text: string, max: number) {
  return text.length > max ? `${text.slice(0, max).trimEnd()}…` : text
}
