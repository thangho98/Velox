import { useTranslation } from 'react-i18next'
import { usePreferences, useUpdatePreferences } from '@/hooks/stores/useAuth'
import { LuGlobe, LuCheck } from 'react-icons/lu'
import { useState, useRef, useEffect } from 'react'

const LANGUAGES = [
  { code: 'en', label: 'English', flag: '🇺🇸' },
  { code: 'vi', label: 'Tiếng Việt', flag: '🇻🇳' },
]

interface LanguageSwitcherProps {
  compact?: boolean
}

export function LanguageSwitcher({ compact }: LanguageSwitcherProps) {
  const { i18n } = useTranslation()
  const { data: preferences } = usePreferences()
  const { mutate: updatePreferences } = useUpdatePreferences()
  const [isOpen, setIsOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  const currentLang = i18n.language || 'en'
  const currentLanguage = LANGUAGES.find((l) => l.code === currentLang) || LANGUAGES[0]

  // Close on click outside
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setIsOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const handleChange = (lang: string) => {
    // Update i18n
    i18n.changeLanguage(lang)

    // Persist to localStorage
    localStorage.setItem('velox-language', lang)

    // Sync to backend if logged in
    if (preferences) {
      updatePreferences({ ...preferences, language: lang })
    }

    setIsOpen(false)
  }

  if (compact) {
    return (
      <div ref={ref} className="relative">
        <button
          onClick={() => setIsOpen(!isOpen)}
          className="flex items-center gap-2 px-4 py-2 text-sm text-gray-300 hover:text-white hover:bg-netflix-gray w-full text-left"
        >
          <LuGlobe size={16} />
          <span>{currentLanguage.flag}</span>
          <span>{currentLanguage.label}</span>
        </button>

        {isOpen && (
          <div className="absolute right-0 mt-1 w-40 rounded bg-netflix-dark border border-netflix-gray shadow-xl z-50">
            {LANGUAGES.map((lang) => (
              <button
                key={lang.code}
                onClick={() => handleChange(lang.code)}
                className={`flex items-center gap-2 w-full px-4 py-2 text-sm text-left hover:bg-netflix-gray ${
                  currentLang === lang.code ? 'text-white' : 'text-gray-400'
                }`}
              >
                <span>{lang.flag}</span>
                <span className="flex-1">{lang.label}</span>
                {currentLang === lang.code && <LuCheck size={14} className="text-netflix-red" />}
              </button>
            ))}
          </div>
        )}
      </div>
    )
  }

  return (
    <div ref={ref} className="relative">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center gap-2 rounded bg-netflix-gray px-4 py-2.5 text-sm text-white hover:bg-gray-600 transition-colors"
      >
        <LuGlobe size={16} className="text-gray-400" />
        <span>{currentLanguage.flag}</span>
        <span>{currentLanguage.label}</span>
      </button>

      {isOpen && (
        <div className="absolute left-0 top-full mt-2 w-48 rounded bg-netflix-dark border border-netflix-gray shadow-xl z-50">
          {LANGUAGES.map((lang) => (
            <button
              key={lang.code}
              onClick={() => handleChange(lang.code)}
              className={`flex items-center gap-3 w-full px-4 py-3 text-sm text-left hover:bg-netflix-gray transition-colors ${
                currentLang === lang.code ? 'text-white bg-netflix-gray/50' : 'text-gray-400'
              }`}
            >
              <span className="text-lg">{lang.flag}</span>
              <span className="flex-1">{lang.label}</span>
              {currentLang === lang.code && <LuCheck size={16} className="text-netflix-red" />}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
