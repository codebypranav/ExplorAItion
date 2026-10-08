"use client"

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import React, { useEffect, useState } from 'react'

import { API_BASE, health } from '../lib/api'

const LINKS = [
  { href: '/', label: 'Home' },
  { href: '/search', label: 'Search' },
  { href: '/itinerary', label: 'Itinerary' },
]

export default function Header() {
  const pathname = usePathname()
  const [state, setState] = useState<'checking' | 'up' | 'down'>('checking')

  useEffect(() => {
    const controller = new AbortController()
    let cancelled = false

    const check = async () => {
      const ok = await health(controller.signal)
      if (!cancelled) setState(ok ? 'up' : 'down')
    }
    check()
    const timer = setInterval(check, 30_000)

    return () => {
      cancelled = true
      controller.abort()
      clearInterval(timer)
    }
  }, [])

  const label = {
    checking: 'checking API…',
    up: 'API online',
    down: 'API offline',
  }[state]

  return (
    <header className="site-header">
      <div className="shell">
        <Link href="/" className="brand">
          <span className="brand-mark" aria-hidden="true">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M3 19l6-14 6 9 3-5 3 10z" />
            </svg>
          </span>
          <span>
            Explor<mark>AI</mark>tion
          </span>
        </Link>

        <nav className="site-nav" aria-label="Main">
          {LINKS.map((link) => (
            <Link
              key={link.href}
              href={link.href}
              className="nav-link"
              aria-current={pathname === link.href ? 'page' : undefined}
            >
              {link.label}
            </Link>
          ))}
          <span
            className="status-dot"
            data-state={state}
            title={`${label} — ${API_BASE}`}
          >
            <i aria-hidden="true" />
            <span>{label}</span>
          </span>
        </nav>
      </div>
    </header>
  )
}
