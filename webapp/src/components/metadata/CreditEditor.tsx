import { LuPlus, LuTrash2 } from 'react-icons/lu'
import type { CreditInput } from '@/types/api'

interface CreditEditorProps {
  credits: CreditInput[]
  onChange: (credits: CreditInput[]) => void
}

export function CreditEditor({ credits, onChange }: CreditEditorProps) {
  const cast = credits.filter((c) => c.role === 'cast')
  const crew = credits.filter((c) => c.role !== 'cast')

  function updateCredit(index: number, field: keyof CreditInput, value: string | number) {
    const updated = [...credits]
    updated[index] = { ...updated[index], [field]: value }
    onChange(updated)
  }

  function addCast() {
    onChange([...credits, { person_name: '', character: '', role: 'cast', order: cast.length }])
  }

  function addCrew(role: 'director' | 'writer') {
    onChange([...credits, { person_name: '', role, order: 0 }])
  }

  function removeCredit(index: number) {
    onChange(credits.filter((_, i) => i !== index))
  }

  function globalIndex(role: string, localIdx: number): number {
    let count = 0
    for (let i = 0; i < credits.length; i++) {
      if (role === 'cast' ? credits[i].role === 'cast' : credits[i].role !== 'cast') {
        if (count === localIdx) return i
        count++
      }
    }
    return -1
  }

  return (
    <div className="space-y-4">
      {/* Cast */}
      <div>
        <div className="mb-2 flex items-center justify-between">
          <label className="text-sm font-medium text-gray-300">Cast</label>
          <button
            onClick={addCast}
            className="flex items-center gap-1 text-xs text-blue-400 hover:text-blue-300"
          >
            <LuPlus size={14} /> Add
          </button>
        </div>
        <div className="space-y-2">
          {cast.map((c, i) => {
            const gi = globalIndex('cast', i)
            return (
              <div key={gi} className="flex items-center gap-2">
                <input
                  type="text"
                  value={c.person_name}
                  onChange={(e) => updateCredit(gi, 'person_name', e.target.value)}
                  placeholder="Name"
                  className="flex-1 rounded-lg bg-[#2a2a2a] px-3 py-1.5 text-sm text-white outline-none focus:ring-1 focus:ring-blue-500"
                />
                <span className="text-gray-500">→</span>
                <input
                  type="text"
                  value={c.character ?? ''}
                  onChange={(e) => updateCredit(gi, 'character', e.target.value)}
                  placeholder="Character"
                  className="flex-1 rounded-lg bg-[#2a2a2a] px-3 py-1.5 text-sm text-white outline-none focus:ring-1 focus:ring-blue-500"
                />
                <button
                  onClick={() => removeCredit(gi)}
                  className="text-gray-500 hover:text-red-400"
                >
                  <LuTrash2 size={16} />
                </button>
              </div>
            )
          })}
        </div>
      </div>

      {/* Crew */}
      <div>
        <div className="mb-2 flex items-center justify-between">
          <label className="text-sm font-medium text-gray-300">Crew</label>
          <div className="flex gap-2">
            <button
              onClick={() => addCrew('director')}
              className="text-xs text-blue-400 hover:text-blue-300"
            >
              + Director
            </button>
            <button
              onClick={() => addCrew('writer')}
              className="text-xs text-blue-400 hover:text-blue-300"
            >
              + Writer
            </button>
          </div>
        </div>
        <div className="space-y-2">
          {crew.map((c, i) => {
            const gi = globalIndex('crew', i)
            return (
              <div key={gi} className="flex items-center gap-2">
                <input
                  type="text"
                  value={c.person_name}
                  onChange={(e) => updateCredit(gi, 'person_name', e.target.value)}
                  placeholder="Name"
                  className="flex-1 rounded-lg bg-[#2a2a2a] px-3 py-1.5 text-sm text-white outline-none focus:ring-1 focus:ring-blue-500"
                />
                <select
                  value={c.role}
                  onChange={(e) => updateCredit(gi, 'role', e.target.value)}
                  className="rounded-lg bg-[#2a2a2a] px-2 py-1.5 text-sm text-white outline-none"
                >
                  <option value="director">Director</option>
                  <option value="writer">Writer</option>
                </select>
                <button
                  onClick={() => removeCredit(gi)}
                  className="text-gray-500 hover:text-red-400"
                >
                  <LuTrash2 size={16} />
                </button>
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}
