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
    const q = new URLSearchParams(location.search).get('q')
    if (location.pathname === '/search' && !q) {
      setSearchQuery('')
    } else if (location.pathname === '/search' && q) {
      setSearchQuery(q)
    }
  }, [location.pathname, location.search])

  return (
    <header
      className={`fixed top-0 left-0 right-0 z-50 transition-all duration-300 ${
        scrolled || isSearchOpen
          ? 'bg-netflix-black/95 backdrop-blur-md'
          : 'bg-gradient-to-b from-black/70 to-transparent'
      }`}
    >
      <div className="flex h-16 items-center justify-between gap-3 px-3 sm:px-4 lg:px-8">
        {/* Left: Logo + Nav */}
        <div className="flex min-w-0 items-center gap-3 sm:gap-8">
          <Logo size="sm" className="shrink-0 sm:text-2xl" />
          {isAuthenticated && (
            <nav className="hidden items-center gap-6 md:flex">
              {navItems.map((item) => (
                <Link
                  key={item.path}
                  to={item.path}
                  className={`text-sm font-medium transition-colors hover:text-netflix-white ${
                    location.pathname === item.path
                      ? 'text-netflix-white font-semibold'
                      : 'text-gray-300'
                  }`}
                >
                  {t(item.labelKey)}
                </Link>
              ))}
            </nav>
          )}
        </div>

        {/* Right: Search + User */}
        <div className="flex shrink-0 items-center gap-2 sm:gap-4">
          {isAuthenticated && (
            <div className="flex items-center gap-1 sm:gap-2">
              <NotificationBell />
              <div className="relative">
                {isSearchOpen ? (
                  <form
                    onSubmit={handleSearch}
                    className="flex items-center gap-2 rounded border border-gray-500 bg-netflix-black/80 px-2.5 py-1.5 sm:px-3"
                  >
                    <LuSearch size={16} className="text-gray-400" />
                    <input
                      ref={searchInputRef}
                      type="text"
                      placeholder={t('search.placeholder')}
                      value={searchQuery}
                      onChange={(e) => setSearchQuery(e.target.value)}
                      className="w-28 min-w-0 bg-transparent text-sm text-white placeholder-gray-400 outline-none sm:w-48"
                    />
                    <button
                      type="button"
                      onClick={toggleSearch}
                      className="text-gray-400 hover:text-white"
                    >
                      <LuX size={16} />
                    </button>
                  </form>
                ) : (
                  <button
                    onClick={toggleSearch}
                    className="p-1.5 text-gray-300 transition-colors hover:text-white sm:p-2"
                    aria-label="Search"
                  >
                    <LuSearch className="h-4 w-4 sm:h-5 sm:w-5" />
                  </button>
                )}
              </div>
            </div>
          )}

          {isAuthenticated && user && (
            <div ref={userMenuRef} className="relative">
              <button
                onClick={() => setIsUserMenuOpen(!isUserMenuOpen)}
                className="flex items-center gap-1.5 text-sm font-medium text-white hover:opacity-80 sm:gap-2"
              >
                <div className="flex h-7 w-7 items-center justify-center rounded bg-netflix-red text-xs font-bold sm:h-8 sm:w-8 sm:text-sm">
                  {user.display_name?.[0]?.toUpperCase() ||
                    user.username?.[0]?.toUpperCase() ||
                    'U'}
                </div>
                <LuChevronDown
                  size={12}
                  className={`hidden transition-transform sm:block ${isUserMenuOpen ? 'rotate-180' : ''}`}
                />
              </button>

              {isUserMenuOpen && (
                <div className="absolute right-0 top-full mt-2 w-48 rounded bg-netflix-dark border border-netflix-gray shadow-xl">
                  <div className="px-4 py-3 border-b border-netflix-gray">
                    <p className="text-sm font-medium text-white">
                      {user.display_name || user.username}
                    </p>
                    <p className="text-xs text-gray-400">
                      {user.is_admin ? t('userMenu.admin') : t('userMenu.user')}
                    </p>
                  </div>
                  <div className="py-1">
                    <LanguageSwitcher compact />
                    <Link
                      to="/settings"
                      className="flex items-center gap-2 px-4 py-2 text-sm text-gray-300 hover:bg-netflix-gray hover:text-white"
                      onClick={() => setIsUserMenuOpen(false)}
                    >
                      <LuSettings size={16} />
                      {t('userMenu.settings')}
                    </Link>
                    <div className="my-1 border-t border-netflix-gray" />
                    <button
                      onClick={() => logout()}
                      className="flex w-full items-center gap-2 px-4 py-2 text-sm text-gray-300 hover:bg-netflix-gray hover:text-white"
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
    </header>
  )
}
