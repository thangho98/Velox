import { Link } from 'react-router'

interface LogoProps {
  className?: string
  size?: 'sm' | 'md' | 'lg'
}

export function Logo({ className = '', size = 'md' }: LogoProps) {
  const sizeClasses = {
    sm: 'text-xl',
    md: 'text-2xl',
    lg: 'text-3xl',
  }

  return (
    <Link
      to="/"
      className={`font-bold tracking-tight text-netflix-red hover:text-netflix-red-hover transition-colors ${sizeClasses[size]} ${className}`}
    >
      VELOX
    </Link>
  )
}
