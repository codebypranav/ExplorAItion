import React from 'react'

import type { Weather } from '../lib/types'
import { describeWeather } from '../lib/weather'

export default function WeatherChip({ weather }: { weather: Weather }) {
  const w = describeWeather(weather)
  return (
    <span className="chip" title={`${w.label} · wind ${w.wind} kph`}>
      <span aria-hidden="true">{w.icon}</span>
      <span className="mono">{w.celsius}°C</span>
    </span>
  )
}
