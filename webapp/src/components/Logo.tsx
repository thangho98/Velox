import { Link } from 'react-router'

interface LogoProps {
  className?: string
  size?: 'sm' | 'md' | 'lg'
}

export function Logo({ className = '', size = 'md' }: LogoProps) {
  const sizeClasses = {
    sm: 'text-[22px]',
    md: 'text-[22px]',
    lg: 'text-[28px]',
  }

  return (
    <Link
      to="/"
      className={`font-extrabold uppercase tracking-[0.22em] text-crimson-500 transition-opacity hover:opacity-90 ${sizeClasses[size]} ${className}`}
    >
      VELOX
    </Link>
  )
}
