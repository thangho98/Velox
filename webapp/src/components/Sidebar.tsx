import { Link, useLocation } from 'react-router'
import { LuHouse, LuFilm, LuTv, LuHeart, LuClock, LuSettings, LuFolderOpen } from 'react-icons/lu'
import { useAuthStore } from '@/stores/auth'
import { useUIStore } from '@/stores/ui'

const sidebarItems = [
  { label: 'Home', path: '/', icon: LuHouse },
  { label: 'Movies', path: '/movies', icon: LuFilm },
  { label: 'Series', path: '/series', icon: LuTv },
  { label: 'Browse', path: '/browse', icon: LuFolderOpen },
  { label: 'Favorites', path: '/favorites', icon: LuHeart },
  { label: 'Recently Watched', path: '/recently-watched', icon: LuClock },
  { label: 'Settings', path: '/settings', icon: LuSettings },
]

export function Sidebar() {
  const location = useLocation()
  const { isAuthenticated } = useAuthStore()
  const { sidebarCollapsed } = useUIStore()

  if (!isAuthenticated) return null

  return (
    <>
      {/* Desktop Sidebar */}
      <aside
        className={`fixed left-0 top-16 bottom-0 z-40 hidden bg-netflix-black/95 backdrop-blur-md border-r border-netflix-gray/50 transition-all duration-300 lg:block ${
          sidebarCollapsed ? 'w-16' : 'w-56'
        }`}
      >
        <nav className="flex flex-col gap-1 p-2">
          {sidebarItems.map((item) => {
            const isActive = location.pathname === item.path
            const Icon = item.icon
            return (
              <Link
                key={item.path}
                to={item.path}
                className={`flex items-center gap-3 rounded px-3 py-3 text-sm font-medium transition-colors ${
                  isActive
                    ? 'bg-netflix-gray text-white'
                    : 'text-gray-400 hover:bg-netflix-gray/50 hover:text-white'
                } ${sidebarCollapsed ? 'justify-center' : ''}`}
                title={sidebarCollapsed ? item.label : undefined}
              >
                <Icon size={20} className="flex-shrink-0" />
                {!sidebarCollapsed && <span>{item.label}</span>}
              </Link>
            )
          })}
        </nav>
      </aside>

      {/* Mobile Bottom Navigation */}
      <nav className="fixed bottom-0 left-0 right-0 z-40 flex items-center justify-around border-t border-netflix-gray/50 bg-netflix-black/95 backdrop-blur-md lg:hidden">
        {sidebarItems.slice(0, 5).map((item) => {
          const isActive = location.pathname === item.path
          const Icon = item.icon
          return (
            <Link
              key={item.path}
              to={item.path}
              className={`flex flex-col items-center gap-1 px-4 py-2 text-xs font-medium transition-colors ${
                isActive ? 'text-white' : 'text-gray-400 hover:text-white'
              }`}
            >
              <Icon size={20} />
              <span>{item.label}</span>
            </Link>
          )
        })}
      </nav>
    </>
  )
}
