import { type SelectHTMLAttributes } from 'react'

// Inline SVG chevron for custom dropdown arrow
const chevronBg = `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16' viewBox='0 0 24 24' fill='none' stroke='%239ca3af' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='m6 9 6 6 6-6'/%3E%3C/svg%3E")`

type SelectSize = 'sm' | 'md'

interface SelectProps extends Omit<SelectHTMLAttributes<HTMLSelectElement>, 'size'> {
  size?: SelectSize
}

const sizeClasses: Record<SelectSize, string> = {
  sm: 'text-xs pl-2.5 pr-7 py-1.5',
  md: 'text-sm pl-3 pr-8 py-2',
}

export function Select({ size = 'md', className = '', style, ...props }: SelectProps) {
  return (
    <select
      className={`appearance-none rounded-md bg-[#0f0f0f] border border-gray-700 text-white focus:outline-none focus:border-[#e50914] transition-colors cursor-pointer bg-[length:16px_16px] bg-no-repeat ${sizeClasses[size]} ${className}`}
      style={{
        backgroundImage: chevronBg,
        backgroundPosition: size === 'sm' ? 'right 6px center' : 'right 8px center',
        ...style,
      }}
      {...props}
    />
  )
}
