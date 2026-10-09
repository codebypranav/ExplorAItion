import type { Metadata } from 'next'
import React, { Suspense } from 'react'

import { PlaceListSkeleton } from '../../components/Skeletons'
import ItineraryClient from './ItineraryClient'

export const metadata: Metadata = {
  title: 'Itinerary planner',
  description:
    'Turn a city, a number of days and a few interests into a routed day-by-day plan.',
}

export default function ItineraryPage() {
  return (
    <Suspense
      fallback={
        <div className="shell page">
          <PlaceListSkeleton count={3} />
        </div>
      }
    >
      <ItineraryClient />
    </Suspense>
  )
}
