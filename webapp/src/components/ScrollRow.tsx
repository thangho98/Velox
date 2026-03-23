import { useRef, useState, useCallback, useEffect, type ReactNode } from 'react'
import { LuChevronLeft, LuChevronRight } from 'react-icons/lu'

interface ScrollRowProps {
  children: ReactNode
}

export function ScrollRow({ children }: ScrollRowProps) {
  const scrollRef = useRef<HTMLDivElement>(null)
  const [canScrollLeft, setCanScrollLeft] = useState(false)
  const [canScrollRight, setCanScrollRight] = useState(false)
  const rafRef = useRef<number | null>(null)

  const updateScrollState = useCallback(() => {
    const el = scrollRef.current
    if (!el) return
    setCanScrollLeft(el.scrollLeft > 0)
    setCanScrollRight(el.scrollLeft + el.clientWidth < el.scrollWidth - 1)
  }, [])

  // Throttle scroll events to one update per animation frame
  const handleScroll = useCallback(() => {
    if (rafRef.current !== null) return
    rafRef.current = requestAnimationFrame(() => {
      updateScrollState()
      rafRef.current = null
    })
  }, [updateScrollState])

  useEffect(() => {
    const id = requestAnimationFrame(updateScrollState)
    return () => {
      cancelAnimationFrame(id)
      if (rafRef.current !== null) cancelAnimationFrame(rafRef.current)
    }
  }, [children, updateScrollState])

  const scroll = useCallback((direction: 'left' | 'right') => {
    const el = scrollRef.current
    if (!el) return
    const amount = el.clientWidth * 0.75
    el.scrollBy({ left: direction === 'right' ? amount : -amount, behavior: 'smooth' })
  }, [])

  return (
    <div className="group/scroll relative">
      {canScrollLeft && (
        <button
          onClick={() => scroll('left')}
          className="absolute left-2 top-1/2 z-10 -translate-y-1/2 flex h-10 w-10 items-center justify-center rounded-full bg-black/70 text-white opacity-0 transition-all group-hover/scroll:opacity-100 hover:bg-black/90 hover:scale-110"
          aria-label="Scroll left"
        >
          <LuChevronLeft size={22} />
        </button>
      )}

      {canScrollRight && (
        <button
          onClick={() => scroll('right')}
          className="absolute right-2 top-1/2 z-10 -translate-y-1/2 flex h-10 w-10 items-center justify-center rounded-full bg-black/70 text-white opacity-0 transition-all group-hover/scroll:opacity-100 hover:bg-black/90 hover:scale-110"
          aria-label="Scroll right"
        >
          <LuChevronRight size={22} />
        </button>
      )}

      <div
        ref={scrollRef}
        onScroll={handleScroll}
        className="hide-scrollbar flex gap-3 overflow-x-auto pb-2"
      >
        {children}
      </div>
    </div>
  )
}
