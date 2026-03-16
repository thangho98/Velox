import { useState, useRef, useEffect, useLayoutEffect } from 'react'
import { createPortal } from 'react-dom'
import { LuEllipsisVertical } from 'react-icons/lu'

export interface ActionMenuItem {
  label: string
  icon?: React.ReactNode
  onClick: () => void
  disabled?: boolean
  danger?: boolean
  adminOnly?: boolean
  separator?: boolean
}

interface ActionMenuProps {
  items: ActionMenuItem[]
  isAdmin?: boolean
  size?: number
}

export function ActionMenu({ items, isAdmin, size = 20 }: ActionMenuProps) {
  const [open, setOpen] = useState(false)
  const buttonRef = useRef<HTMLButtonElement>(null)
  const menuRef = useRef<HTMLDivElement>(null)
  const [pos, setPos] = useState({ top: 0, left: 0 })

  useLayoutEffect(() => {
    if (!open || !buttonRef.current) return
    const rect = buttonRef.current.getBoundingClientRect()
    setPos({
      top: rect.bottom + 4,
      left: rect.right,
    })
  }, [open])

  useEffect(() => {
    if (!open) return
    function handleClick(e: MouseEvent) {
      if (
        menuRef.current &&
        !menuRef.current.contains(e.target as Node) &&
        buttonRef.current &&
        !buttonRef.current.contains(e.target as Node)
      ) {
        setOpen(false)
      }
    }
    function handleScroll() {
      setOpen(false)
    }
    document.addEventListener('mousedown', handleClick)
    window.addEventListener('scroll', handleScroll, true)
    return () => {
      document.removeEventListener('mousedown', handleClick)
      window.removeEventListener('scroll', handleScroll, true)
    }
  }, [open])

  const visibleItems = items.filter((item) => !item.adminOnly || isAdmin)
  if (visibleItems.length === 0) return null

  return (
    <>
      <button
        ref={buttonRef}
        onClick={(e) => {
          e.preventDefault()
          e.stopPropagation()
          setOpen(!open)
        }}
        className="rounded-full p-1.5 text-gray-400 transition-colors hover:bg-white/10 hover:text-white"
      >
        <LuEllipsisVertical size={size} />
      </button>

      {open &&
        createPortal(
          <div
            ref={menuRef}
            className="fixed z-[9999] min-w-[180px] overflow-hidden rounded-lg bg-[#2a2a2a] py-1 shadow-xl ring-1 ring-white/10"
            style={{ top: pos.top, left: pos.left, transform: 'translateX(-100%)' }}
          >
            {visibleItems.map((item, i) => (
              <div key={i}>
                {item.separator && i > 0 && <div className="my-1 border-t border-gray-700" />}
                <button
                  onClick={(e) => {
                    e.preventDefault()
                    e.stopPropagation()
                    item.onClick()
                    setOpen(false)
                  }}
                  disabled={item.disabled}
                  className={`flex w-full items-center gap-3 px-4 py-2 text-left text-sm transition-colors disabled:opacity-40 ${
                    item.danger
                      ? 'text-red-400 hover:bg-red-500/10'
                      : 'text-gray-200 hover:bg-white/10'
                  }`}
                >
                  {item.icon && <span className="flex-shrink-0">{item.icon}</span>}
                  {item.label}
                </button>
              </div>
            ))}
          </div>,
          document.body,
        )}
    </>
  )
}
