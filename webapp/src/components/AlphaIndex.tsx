/* eslint-disable react-refresh/only-export-components */
import { useState, useEffect, useCallback } from 'react'

interface AlphaIndexProps {
  activeLetters: Set<string>
  onSelect: (letter: string) => void
  currentLetter: string
}

const ALPHABET = '#ABCDEFGHIJKLMNOPQRSTUVWXYZ'.split('')

export function AlphaIndex({ activeLetters, onSelect, currentLetter }: AlphaIndexProps) {
  return (
    <nav className="fixed right-2 top-1/2 z-30 -translate-y-1/2 flex flex-col items-center gap-0.5">
      {ALPHABET.map((letter) => {
        const isActive = activeLetters.has(letter)
        return (
          <button
            key={letter}
            onClick={() => isActive && onSelect(letter)}
            className={`w-5 text-center text-[11px] leading-4 transition-all rounded-sm ${
              currentLetter === letter
                ? 'bg-[#e50914] text-white font-bold scale-125'
                : isActive
                  ? 'text-white hover:text-[#e50914] cursor-pointer'
                  : 'text-gray-600/50 cursor-default'
            }`}
          >
            {letter}
          </button>
        )
      })}
    </nav>
  )
}

function getLetterForTitle(title: string): string {
  const first = title.charAt(0).toUpperCase()
  return /[A-Z]/.test(first) ? first : '#'
}

/**
 * Hook for alpha index scroll tracking.
 * Each card wrapper should have data-alpha-letter on the FIRST item of each letter group.
 */
export function useAlphaScroll(items: { sort_title?: string; title: string }[] | undefined) {
  const [currentLetter, setCurrentLetter] = useState('')

  const handleScroll = useCallback(() => {
    const markers = document.querySelectorAll<HTMLElement>('[data-alpha-letter]')
    let current = ''
    for (const el of markers) {
      if (el.getBoundingClientRect().top <= 200) {
        current = el.dataset.alphaLetter || ''
      } else {
        break
      }
    }
    if (current) setCurrentLetter(current)
  }, [])

  useEffect(() => {
    window.addEventListener('scroll', handleScroll, { passive: true })
    requestAnimationFrame(handleScroll)
    return () => window.removeEventListener('scroll', handleScroll)
  }, [handleScroll])

  return { currentLetter, getLetterForTitle }
}
