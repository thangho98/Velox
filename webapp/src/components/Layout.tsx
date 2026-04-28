import type { ReactNode } from 'react'
import { MobileTabBar } from './MobileTabBar'
import { Navbar } from './Navbar'

interface LayoutProps {
  children: ReactNode
  fullWidth?: boolean
}

export function Layout({ children, fullWidth = false }: LayoutProps) {
  return (
    <div className="min-h-screen bg-ink-0">
      <Navbar />
      <MobileTabBar />

      <main className={`min-h-screen ${fullWidth ? 'pt-0' : 'pt-[72px]'}`}>
        <div
          className={fullWidth ? '' : 'mx-auto max-w-[1600px] px-4 pb-28 sm:px-6 lg:px-10 lg:pb-10'}
        >
          {children}
        </div>
      </main>
    </div>
  )
}
