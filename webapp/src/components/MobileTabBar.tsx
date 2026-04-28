import { Link, useLocation } from 'react-router'
import { LuHouse, LuFilm, LuTv, LuFolderOpen, LuMonitorPlay } from 'react-icons/lu'
import { useAuthStore } from '@/stores/auth'
import { useTranslation } from '@/hooks/useTranslation'

const mobileItems = [
  { labelKey: 'nav.home', path: '/', icon: LuHouse },
  { labelKey: 'nav.movies', path: '/movies', icon: LuFilm },
  { labelKey: 'nav.series', path: '/series', icon: LuTv },
  { labelKey: 'nav.liveTv', path: '/livetv', icon: LuMonitorPlay, isLive: true },
  { labelKey: 'nav.browse', path: '/browse', icon: LuFolderOpen },
]

export function MobileTabBar() {
  const location = useLocation()
  const { isAuthenticated } = useAuthStore()
  const { t } = useTranslation('navigation')

  if (!isAuthenticated) return null

  return (
    <nav className="fixed inset-x-0 bottom-0 z-50 border-t border-white/5 bg-black/95 px-2 pb-[max(env(safe-area-inset-bottom),0.75rem)] pt-2 backdrop-blur-md lg:hidden">
      <div className="mx-auto grid max-w-md grid-cols-5 gap-1 md:max-w-3xl">
        {mobileItems.map((item) => {
          const isActive = location.pathname === item.path
          const Icon = item.icon

          return (
            <Link
              key={item.path}
              to={item.path}
              aria-current={isActive ? 'page' : undefined}
              className={`relative flex flex-col items-center gap-1 rounded-sm px-2 py-2 text-[11px] font-medium transition-colors ${
                isActive ? 'text-white' : 'text-fog-500'
              }`}
            >
              <div className="relative">
                <Icon size={19} />
                {item.isLive && (
                  <span className="absolute -right-1 -top-1 h-2.5 w-2.5 rounded-full bg-crimson-500">
                    <span className="live-dot absolute inset-0 rounded-full bg-crimson-500" />
                  </span>
                )}
              </div>
              <span>{t(item.labelKey)}</span>
            </Link>
          )
        })}
      </div>
    </nav>
  )
}
