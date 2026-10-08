import type { Metadata } from 'next'
import React, { Suspense } from 'react'

import { PlaceListSkeleton } from '../../components/Skeletons'
import SearchClient from './SearchClient'

export const metadata: Metadata = {
  title: 'Search places',
  description:
    'Rank points of interest by how well they match a plain-English request.',
}

export default function SearchPage() {
  return (
    <Suspense
      fallback={
        <div className="shell page">
          <PlaceListSkeleton count={3} />
        </div>
      }
    >
      <SearchClient />
    </Suspense>
  )
}
