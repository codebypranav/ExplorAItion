"use client"

import React from 'react'

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  return (
    <div className="shell page">
      <div className="notice notice-error" role="alert">
        <span aria-hidden="true">⚠️</span>
        <div>
          <strong>The page hit an unexpected error.</strong>
          <div className="small mono" style={{ marginTop: 6 }}>
            {error.message}
          </div>
        </div>
        <div className="notice-actions">
          <button type="button" className="btn btn-sm" onClick={reset}>
            Reload
          </button>
        </div>
      </div>
    </div>
  )
}
