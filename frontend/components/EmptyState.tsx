import React from 'react'

export default function EmptyState({
  emoji = '🧭',
  title,
  children,
}: {
  emoji?: string
  title: string
  children?: React.ReactNode
}) {
  return (
    <div className="empty">
      <span className="emoji" aria-hidden="true">
        {emoji}
      </span>
      <h3>{title}</h3>
      {children ? <div className="small muted">{children}</div> : null}
    </div>
  )
}
