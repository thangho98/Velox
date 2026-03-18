import { Link, useNavigate, useLocation } from 'react-router'
import { useState, useRef, useEffect } from 'react'
import { LuSearch, LuX, LuChevronDown, LuLogOut, LuSettings } from 'react-icons/lu'
import { Logo } from './Logo'
import { useAuthStore } from '@/stores/auth'
import { useLogout } from '@/hooks/stores/useAuth'
import { useUIStore } from '@/stores/ui'

const navItems = [
  { label: 'Home', path: '/' },
  { label: 'Movies', path: '/movies' },
  { label: 'Series', path: '/series' },
  { label: 'Browse', path: '/browse' },
]

export function Navbar() {
  const navigate = useNavigate()
  const location = useLocation()
  const { user, isAuthenticated } = useAuthStore()
  const { mutate: logout } = useLogout()
  const { isSearchOpen, toggleSearch } = useUIStore()

  const [scrolled, setScrolled] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [isUserMenuOpen, setIsUserMenuOpen] = useState(false)
  const userMenuRef = useRef<HTMLDivElement>(null)
  const searchInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    const handleScroll = () => setScrolled(window.scrollY > 0)
    window.addEventListener('scroll', handleScroll)
    return () => window.removeEventListener('scroll', handleScroll)
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

  return (
    <header
      className={`fixed top-0 left-0 right-0 z-50 transition-all duration-300 ${
        scrolled || isSearchOpen
          ? 'bg-netflix-black/95 backdrop-blur-md'
          : 'bg-gradient-to-b from-black/70 to-transparent'
      }`}
    >
      <div className="flex h-16 items-center justify-between px-4 lg:px-8">
        {/* Left: Logo + Nav */}
        <div className="flex items-center gap-8">
          <Logo />
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
                  {item.label}
                </Link>
              ))}
            </nav>
          )}
        </div>

        {/* Right: Search + User */}
        <div className="flex items-center gap-4">
          {isAuthenticated && (
            <div className="relative">
              {isSearchOpen ? (
                <form
                  onSubmit={handleSearch}
                  className="flex items-center gap-2 rounded bg-netflix-black/80 border border-gray-500 px-3 py-1.5"
                >
                  <LuSearch size={16} className="text-gray-400" />
                  <input
                    ref={searchInputRef}
                    type="text"
                    placeholder="Search movies, genres..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="w-48 bg-transparent text-sm text-white placeholder-gray-400 outline-none"
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
                  className="p-2 text-gray-300 transition-colors hover:text-white"
                  aria-label="Search"
                >
                  <LuSearch size={20} />
                </button>
              )}
            </div>
          )}

          {isAuthenticated && user && (
            <div ref={userMenuRef} className="relative">
              <button
                onClick={() => setIsUserMenuOpen(!isUserMenuOpen)}
                className="flex items-center gap-2 text-sm font-medium text-white hover:opacity-80"
              >
                <div className="flex h-8 w-8 items-center justify-center rounded bg-netflix-red text-sm font-bold">
                  {user.display_name?.[0]?.toUpperCase() ||
                    user.username?.[0]?.toUpperCase() ||
                    'U'}
                </div>
                <LuChevronDown
                  size={12}
                  className={`transition-transform ${isUserMenuOpen ? 'rotate-180' : ''}`}
                />
              </button>

              {isUserMenuOpen && (
                <div className="absolute right-0 top-full mt-2 w-48 rounded bg-netflix-dark border border-netflix-gray shadow-xl">
                  <div className="px-4 py-3 border-b border-netflix-gray">
                    <p className="text-sm font-medium text-white">
                      {user.display_name || user.username}
                    </p>
                    <p className="text-xs text-gray-400">
                      {user.is_admin ? 'Administrator' : 'User'}
                    </p>
                  </div>
                  <div className="py-1">
                    <Link
                      to="/settings"
                      className="flex items-center gap-2 px-4 py-2 text-sm text-gray-300 hover:bg-netflix-gray hover:text-white"
                      onClick={() => setIsUserMenuOpen(false)}
                    >
                      <LuSettings size={16} />
                      Settings
                    </Link>
                    <div className="my-1 border-t border-netflix-gray" />
                    <button
                      onClick={() => logout()}
                      className="flex w-full items-center gap-2 px-4 py-2 text-sm text-gray-300 hover:bg-netflix-gray hover:text-white"
                    >
                      <LuLogOut size={16} />
                      Sign Out
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
