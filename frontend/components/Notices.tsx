"use client"

import React from 'react'

import { API_BASE, ApiError } from '../lib/api'

/** Shown whenever the page is rendering lib/sample.ts data instead of API data. */
export function DemoNotice({ onExit }: { onExit?: () => void }) {
  return (
    <div className="notice notice-info" role="status">
      <span aria-hidden="true">🧪</span>
      <div>
        <strong>Sample data.</strong> These results come from a built-in Paris
        fixture, not from the backend — useful for reviewing the interface while
        the Pinecone index is empty.
      </div>
      {onExit && (
        <div className="notice-actions">
          <button type="button" className="btn btn-sm" onClick={onExit}>
            Use live API
          </button>
        </div>
      )}
    </div>
  )
}

/** Turns an ApiError (or any error) into something actionable. */
export function ErrorNotice({
  error,
  onRetry,
  onUseSample,
}: {
  error: unknown
  onRetry?: () => void
  onUseSample?: () => void
}) {
  const apiError = error instanceof ApiError ? error : null
  const message =
    error instanceof Error ? error.message : 'Something went wrong.'

  return (
    <div className="notice notice-error" role="alert">
      <span aria-hidden="true">⚠️</span>
      <div>
        <strong>
          {apiError?.offline
            ? 'Cannot reach the backend.'
            : `Request failed${apiError?.status ? ` (${apiError.status})` : ''}.`}
        </strong>{' '}
        {message}
        {apiError?.offline && (
          <div className="small" style={{ marginTop: 6 }}>
            Expected at <code className="mono">{API_BASE}</code>. Start it with{' '}
            <code className="mono">./start-dev.sh</code> from the repository root.
          </div>
        )}
      </div>
      <div className="notice-actions">
        {onRetry && (
          <button type="button" className="btn btn-sm" onClick={onRetry}>
            Try again
          </button>
        )}
        {onUseSample && (
          <button type="button" className="btn btn-sm" onClick={onUseSample}>
            Use sample data
          </button>
        )}
      </div>
    </div>
  )
}
