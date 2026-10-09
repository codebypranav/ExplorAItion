import './globals.css'

import type { Metadata, Viewport } from 'next'
import React from 'react'

import Footer from '../components/Footer'
import Header from '../components/Header'

export const metadata: Metadata = {
  title: {
    default: 'ExplorAItion — AI trip planning',
    template: '%s · ExplorAItion',
  },
  description:
    'Describe a trip in plain English and get matching places and a day-by-day route, built with semantic search over a POI index.',
}

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
  themeColor: [
    { media: '(prefers-color-scheme: light)', color: '#faf6ee' },
    { media: '(prefers-color-scheme: dark)', color: '#14120d' },
  ],
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body>
        <Header />
        <main>{children}</main>
        <Footer />
      </body>
    </html>
  )
}
