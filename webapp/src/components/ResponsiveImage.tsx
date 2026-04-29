import { Blurhash } from 'react-blurhash'
import { useState } from 'react'
import type { ImageResource } from '@/types/api'
import { resolveImageUrl } from '@velox/shared/lib/image'

interface Props {
  data: ImageResource | null | undefined
  sizes: string // CSS `sizes` attribute, e.g. "(max-width: 768px) 185px, 500px"
  alt: string
  className?: string
  loading?: 'lazy' | 'eager'
}

export function ResponsiveImage({ data, sizes, alt, className = '', loading = 'lazy' }: Props) {
  const [loaded, setLoaded] = useState(false)
  if (!data) return <div className={`bg-zinc-900 ${className}`} style={{ aspectRatio: '2/3' }} />

  const srcsetAttr = Object.entries(data.srcset)
    .filter(([k]) => k !== 'original')
    .map(([width, url]) => `${resolveImageUrl(url)} ${width}w`)
    .join(', ')

  const aspectStyle = aspectToCss(data.aspect)

  return (
    <div className={`relative overflow-hidden ${className}`} style={aspectStyle}>
      {data.blurhash && !loaded && (
        <Blurhash
          hash={data.blurhash}
          width="100%"
          height="100%"
          resolutionX={32}
          resolutionY={32}
          punch={1}
        />
      )}
      <picture>
        <img
          src={resolveImageUrl(data.url)}
          srcSet={srcsetAttr}
          sizes={sizes}
          alt={alt}
          loading={loading}
          onLoad={() => setLoaded(true)}
          className={`absolute inset-0 h-full w-full object-cover transition-opacity duration-300 ${loaded ? 'opacity-100' : 'opacity-0'}`}
        />
      </picture>
    </div>
  )
}

function aspectToCss(aspect: string): React.CSSProperties {
  if (aspect === '2:3') return { aspectRatio: '2/3' }
  if (aspect === '16:9') return { aspectRatio: '16/9' }
  return {}
}
