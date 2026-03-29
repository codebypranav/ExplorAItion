import './globals.css'
import React from 'react'

export const metadata = {
  title: 'ExplorAItion'
}

export default function RootLayout({ children }){
  return (
    <html>
      <body>
        <div className="mountain-bg"></div>
        {children}
      </body>
    </html>
  )
}
