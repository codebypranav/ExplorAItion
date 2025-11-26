import './globals.css'
import React from 'react'

export const metadata = {
  title: 'ExplorAItion'
}

export default function RootLayout({ children }){
  return (
    <html>
      <body style={{ fontFamily: 'Inter, system-ui, -apple-system, Segoe UI, Roboto', margin: 0 }}>{children}</body>
    </html>
  )
}
