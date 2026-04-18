import { useEffect, useState } from 'react'

export function useIntersectionObserver<T extends Element = HTMLDivElement>(
  options?: IntersectionObserverInit,
) {
  const [isIntersecting, setIsIntersecting] = useState(false)
  const [target, setTarget] = useState<T | null>(null)

  useEffect(() => {
    if (!target) return

    const observer = new IntersectionObserver(([entry]) => {
      setIsIntersecting(entry.isIntersecting)
    }, options)

    observer.observe(target)
    return () => observer.disconnect()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [target, options?.rootMargin, options?.root, options?.threshold])

  return { targetRef: setTarget, isIntersecting }
}
