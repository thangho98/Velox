import { Link, useNavigate, useLocation } from 'react-router'
import { useState, useRef, useEffect } from 'react'
import { LuSearch, LuX, LuChevronDown, LuLogOut, LuSettings } from 'react-icons/lu'
import { Logo } from './Logo'
import { LanguageSwitcher } from './LanguageSwitcher'
import { NotificationBell } from './NotificationBell'
import { useAuthStore } from '@/stores/auth'
import { useLogout } from '@/hooks/stores/useAuth'
import { useTranslation } from '@/hooks/useTranslation'
import { useUIStore } from '@/stores/ui'

const navItems = [
  { labelKey: 'nav.home', path: '/' },
  { labelKey: 'nav.movies', path: '/movies' },
  { labelKey: 'nav.series', path: '/series' },
  { labelKey: 'nav.liveTv', path: '/livetv', isLive: true },
  { labelKey: 'nav.browse', path: '/browse' },
]

export function Navbar() {
  const navigate = useNavigate()
  const location = useLocation()
  const { user, isAuthenticated } = useAuthStore()
  const { mutate: logout } = useLogout()
  const { isSearchOpen, toggleSearch } = useUIStore()
  const { t } = useTranslation('navigation')

  const [scrolled, setScrolled] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [isUserMenuOpen, setIsUserMenuOpen] = useState(false)
  const userMenuRef = useRef<HTMLDivElement>(null)
  const searchInputRef = useRef<HTMLInputElement>(null)
  const scrollRafRef = useRef<number | null>(null)

  useEffect(() => {
    const handleScroll = () => {
      if (scrollRafRef.current !== null) return
      scrollRafRef.current = requestAnimationFrame(() => {
        setScrolled(window.scrollY > 0)
        scrollRafRef.current = null
      })
    }
    window.addEventListener('scroll', handleScroll, { passive: true })
    return () => {
      window.removeEventListener('scroll', handleScroll)
      if (scrollRafRef.current !== null) cancelAnimationFrame(scrollRafRef.current)
    }
  }, [])

  useEffect(() => {
    if (isSearchOpen && searchInputRef.current) {
      searchInputRef.current.focus()
    }
  }, [isSearchOpen])

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (userMenuRef.current && !userMenuRef.current.contains(e.target as Node)) {
        setIsUserMenuOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    if (searchQuery.trim()) {
      navigate(`/search?q=${encodeURIComponent(searchQuery.trim())}`)
      toggleSearch()
    }
  }

  useEffect(() => {
    const nextQuery =
      location.pathname === '/search' ? new URLSearchParams(location.search).get('q') || '' : ''
    queueMicrotask(() => setSearchQuery(nextQuery))
  }, [location.pathname, location.search])

  return (
    <header
      className={`fixed top-0 left-0 right-0 z-50 transition-all duration-300 ${
        scrolled || isSearchOpen
          ? 'topbar bg-black/72 shadow-card'
          : 'bg-gradient-to-b from-black/70 to-transparent'
      }`}
    >
      <div className="mx-auto max-w-[1600px] px-4 sm:px-6 lg:px-10">
        <div className="flex h-[72px] items-center justify-between gap-4">
          <div className="flex min-w-0 items-center gap-8">
            <Logo size="sm" className="shrink-0" />
            {isAuthenticated && (
              <nav className="hidden items-center gap-7 lg:flex">
                {navItems.map((item) => {
                  const isActive = location.pathname === item.path
                  return (
                    <Link
                      key={item.path}
                      to={item.path}
                      aria-current={isActive ? 'page' : undefined}
                      className={`relative inline-flex items-center gap-2 whitespace-nowrap pb-1 text-sm font-medium tracking-[0.02em] transition-colors ${
                        isActive ? 'tab-active text-white' : 'text-fog-500 hover:text-white'
                      }`}
                    >
                      {item.isLive && (
                        <span className="live-dot h-1.5 w-1.5 rounded-full bg-crimson-500" />
                      )}
                      <span>{t(item.labelKey)}</span>
                    </Link>
                  )
                })}
              </nav>
            )}
          </div>

          <div className="flex shrink-0 items-center gap-2 sm:gap-3">
            {isAuthenticated && (
              <div className="flex items-center gap-1 sm:gap-2">
                <NotificationBell />
                <div className="relative">
                  {isSearchOpen ? (
                    <form
                      onSubmit={handleSearch}
                      className="flex items-center gap-2 rounded-sm bg-ink-200/90 px-3 py-2 text-white shadow-card"
                    >
                      <LuSearch size={16} className="text-fog-600" />
                      <input
                        ref={searchInputRef}
                        type="text"
                        placeholder={t('search.placeholder')}
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="w-28 min-w-0 bg-transparent text-sm text-white placeholder:text-fog-500 outline-none sm:w-52"
                      />
                      <button
                        type="button"
                        onClick={toggleSearch}
                        className="text-fog-500 transition-colors hover:text-white"
                      >
                        <LuX size={16} />
                      </button>
                    </form>
                  ) : (
                    <button
                      onClick={toggleSearch}
                      className="flex h-11 w-11 items-center justify-center rounded-sm bg-ink-200/80 text-fog-700 shadow-card transition-colors hover:bg-ink-300 hover:text-white"
                      aria-label="Search"
                    >
                      <LuSearch className="h-[18px] w-[18px]" />
                    </button>
                  )}
                </div>
              </div>
            )}

            {isAuthenticated && user && (
              <div ref={userMenuRef} className="relative">
                <button
                  onClick={() => setIsUserMenuOpen(!isUserMenuOpen)}
                  className="flex items-center gap-2 text-sm font-medium text-white transition-opacity hover:opacity-90"
                >
                  <div className="flex h-11 w-11 items-center justify-center rounded-sm bg-crimson-500 text-sm font-bold shadow-card">
                    {user.display_name?.[0]?.toUpperCase() ||
                      user.username?.[0]?.toUpperCase() ||
                      'U'}
                  </div>
                  <LuChevronDown
                    size={14}
                    className={`hidden text-fog-600 transition-transform sm:block ${isUserMenuOpen ? 'rotate-180' : ''}`}
                  />
                </button>

                {isUserMenuOpen && (
                  <div className="absolute right-0 top-full mt-3 w-52 rounded-sm bg-ink-100 p-2 shadow-pop">
                    <div className="px-3 py-2">
                      <p className="text-sm font-medium text-white">
                        {user.display_name || user.username}
                      </p>
                      <p className="mt-1 text-[11px] uppercase tracking-[0.28em] text-fog-500">
                        {user.is_admin ? t('userMenu.admin') : t('userMenu.user')}
                      </p>
                    </div>
                    <div className="mt-1 space-y-1">
                      <LanguageSwitcher compact />
                      <Link
                        to="/settings"
                        className="flex items-center gap-2 rounded-sm px-3 py-2 text-sm text-fog-700 transition-colors hover:bg-ink-200 hover:text-white"
                        onClick={() => setIsUserMenuOpen(false)}
                      >
                        <LuSettings size={16} />
                        {t('userMenu.settings')}
                      </Link>
                      <button
                        onClick={() => logout()}
                        className="flex w-full items-center gap-2 rounded-sm px-3 py-2 text-sm text-fog-700 transition-colors hover:bg-ink-200 hover:text-white"
                      >
                        <LuLogOut size={16} />
                        {t('userMenu.signOut')}
                      </button>
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </div>
    </header>
  )
}
