import { LuChevronRight, LuHouse, LuFolder } from 'react-icons/lu'

interface BreadcrumbItem {
  name: string
  path: string
  isRoot?: boolean
}

interface BreadcrumbProps {
  items: BreadcrumbItem[]
  onNavigate: (path: string) => void
}

export function Breadcrumb({ items, onNavigate }: BreadcrumbProps) {
  if (items.length === 0) return null

  return (
    <nav className="flex items-center gap-1 text-sm text-gray-400">
      {items.map((item, index) => {
        const isLast = index === items.length - 1

        return (
          <div key={item.path} className="flex items-center">
            {index > 0 && <LuChevronRight className="w-4 h-4 mx-1 text-gray-600" />}

            {isLast ? (
              <span className="flex items-center gap-1.5 text-white font-medium">
                {item.isRoot ? <LuHouse className="w-4 h-4" /> : <LuFolder className="w-4 h-4" />}
                {item.name}
              </span>
            ) : (
              <button
                onClick={() => onNavigate(item.path)}
                className="flex items-center gap-1.5 hover:text-white transition-colors"
              >
                {item.isRoot ? <LuHouse className="w-4 h-4" /> : <LuFolder className="w-4 h-4" />}
                {item.name}
              </button>
            )}
          </div>
        )
      })}
    </nav>
  )
}
