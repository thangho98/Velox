import type { ReactNode } from 'react'
import { Navbar } from './Navbar'
import { Sidebar } from './Sidebar'

interface LayoutProps {
  children: ReactNode
  fullWidth?: boolean
}

export function Layout({ children, fullWidth = false }: LayoutProps) {
  return (
    <div className="min-h-screen bg-netflix-black">
      {/* Fixed Navbar */}
      <Navbar />

      {/* Fixed Sidebar (desktop) — hidden in fullWidth mode (e.g. WatchPage) */}
      {!fullWidth && <Sidebar />}

      {/* Main Content */}
      <main
        className={`min-h-screen pt-16 transition-all duration-300 ${fullWidth ? '' : 'lg:pl-56'}`}
      >
        <div className={fullWidth ? '' : 'p-4 lg:p-8 pb-20 lg:pb-8'}>{children}</div>
      </main>
    </div>
  )
}
