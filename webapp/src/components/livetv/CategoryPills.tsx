import { ScrollRow } from '@/components/ScrollRow'

interface CategoryPillsProps {
  groups: string[]
  selectedGroup: string
  onSelectGroup: (group: string) => void
  className?: string
}

export function CategoryPills({
  groups,
  selectedGroup,
  onSelectGroup,
  className = '',
}: CategoryPillsProps) {
  return (
    <div className={`relative w-full ${className}`}>
      <ScrollRow>
        <button
          type="button"
          onClick={() => onSelectGroup('all')}
          className={`flex-none whitespace-nowrap rounded-sm px-4 py-2 text-sm font-medium transition-colors ${
            selectedGroup === 'all'
              ? 'bg-white text-black'
              : 'bg-ink-200 text-fog-600 hover:bg-ink-300 hover:text-white'
          }`}
        >
          All
        </button>

        {groups.map((group) => {
          if (!group) return null
          const isSelected = selectedGroup === group
          return (
            <button
              key={group}
              type="button"
              onClick={() => onSelectGroup(group)}
              className={`flex-none whitespace-nowrap rounded-sm px-4 py-2 text-sm font-medium transition-colors ${
                isSelected
                  ? 'bg-white text-black'
                  : 'bg-ink-200 text-fog-600 hover:bg-ink-300 hover:text-white'
              }`}
            >
              {group}
            </button>
          )
        })}

        {/* Spacer to push the last item past the right fade mask */}
        <div className="w-8 flex-none sm:w-12 lg:w-16" />
      </ScrollRow>

      {/* Right fade mask — hints scrollability */}
      <div className="pointer-events-none absolute bottom-2 right-0 top-0 w-8 bg-gradient-to-l from-ink-0/95 to-transparent sm:w-12 lg:w-16" />
    </div>
  )
}
