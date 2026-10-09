import Link from 'next/link'
import React from 'react'

export default function NotFound() {
  return (
    <div className="shell page">
      <div className="empty">
        <span className="emoji" aria-hidden="true">
          🧭
        </span>
        <h1 style={{ fontSize: '1.8rem' }}>Off the map</h1>
        <p>That page does not exist.</p>
        <Link href="/" className="btn btn-primary">
          Back to the start
        </Link>
      </div>
    </div>
  )
}
