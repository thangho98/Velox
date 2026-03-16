import { useState, useRef } from 'react'
import { LuX } from 'react-icons/lu'
import { useAllGenres } from '@/hooks/stores/useMedia'

interface GenreEditorProps {
  genres: string[]
  onChange: (genres: string[]) => void
}

export function GenreEditor({ genres, onChange }: GenreEditorProps) {
  const [input, setInput] = useState('')
  const [showSuggestions, setShowSuggestions] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)
  const { data: allGenres = [] } = useAllGenres()

  const suggestions = allGenres
    .map((g) => g.name)
    .filter((name) => !genres.includes(name) && name.toLowerCase().includes(input.toLowerCase()))
    .slice(0, 8)

  function addGenre(name: string) {
    if (name && !genres.includes(name)) {
      onChange([...genres, name])
    }
    setInput('')
    setShowSuggestions(false)
    inputRef.current?.focus()
  }

  function removeGenre(name: string) {
    onChange(genres.filter((g) => g !== name))
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Enter') {
      e.preventDefault()
      if (input.trim()) addGenre(input.trim())
    }
    if (e.key === 'Backspace' && !input && genres.length > 0) {
      removeGenre(genres[genres.length - 1])
    }
  }

  return (
    <div className="space-y-2">
      <label className="text-sm font-medium text-gray-300">Genres</label>
      <div className="flex flex-wrap gap-2">
        {genres.map((g) => (
          <span
            key={g}
            className="flex items-center gap-1 rounded-full bg-blue-600/20 px-3 py-1 text-sm text-blue-300"
          >
            {g}
            <button onClick={() => removeGenre(g)} className="hover:text-white">
              <LuX size={14} />
            </button>
          </span>
        ))}
      </div>
      <div className="relative">
        <input
          ref={inputRef}
          type="text"
          value={input}
          onChange={(e) => {
            setInput(e.target.value)
            setShowSuggestions(true)
          }}
          onFocus={() => setShowSuggestions(true)}
          onBlur={() => setTimeout(() => setShowSuggestions(false), 200)}
          onKeyDown={handleKeyDown}
          placeholder="Type genre name + Enter"
          className="w-full rounded-lg bg-[#2a2a2a] px-3 py-2 text-sm text-white outline-none focus:ring-1 focus:ring-blue-500"
        />
        {showSuggestions && suggestions.length > 0 && (
          <div className="absolute z-10 mt-1 w-full rounded-lg bg-[#333] shadow-lg">
            {suggestions.map((s) => (
              <button
                key={s}
                onMouseDown={() => addGenre(s)}
                className="block w-full px-3 py-2 text-left text-sm text-gray-200 hover:bg-[#444]"
              >
                {s}
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
