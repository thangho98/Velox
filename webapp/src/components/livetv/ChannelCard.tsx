import { useState } from 'react'
import { LuPlay, LuEyeOff, LuEye } from 'react-icons/lu'
import type { LiveChannel } from '@/hooks/stores/useLiveTV'
import { useToggleChannelHidden } from '@/hooks/stores/useLiveTV'
import { ActionMenu } from '@/components/ActionMenu'
import type { ActionMenuItem } from '@/components/ActionMenu'
import { useToast } from '@/components/Toast'

interface ChannelCardProps {
  channel: LiveChannel
  isActive?: boolean
  onClick: () => void
}

const colorMap = [
  'bg-[#12234d]',
  'bg-[#3d141a]',
  'bg-[#2c1c08]',
  'bg-[#083023]',
  'bg-[#112933]',
  'bg-[#1f1540]',
  'bg-[#33152a]',
  'bg-[#1a1a1a]',
]

function getCardColor(channel: LiveChannel) {
  const seed = `${channel.name}:${channel.groupTitle}`
  let hash = 0
  for (let i = 0; i < seed.length; i++) {
    hash = (hash * 31 + seed.charCodeAt(i)) % 997
  }
  return colorMap[Math.abs(hash) % colorMap.length]
}

function stripDecorations(name: string) {
  return name
    .replace(/\s*\([^)]*\)/g, '')
    .replace(/\s*\[[^\]]*\]/g, '')
    .trim()
}

export function ChannelCard({ channel, isActive = false, onClick }: ChannelCardProps) {
  const [imageError, setImageError] = useState(false)
  const bgColor = getCardColor(channel)
  const cleanName = stripDecorations(channel.name) || channel.name

  const { mutate: toggleHidden } = useToggleChannelHidden()
  const { success } = useToast()

  const handleHide = () => {
    const action = channel.isHidden ? 'unhide' : 'hide'
    const doHide = window.confirm(`Are you sure you want to ${action} this channel?`)
    if (doHide) {
      toggleHidden(channel.id, {
        onSuccess: (data) => {
          success(`Channel ${channel.name} is now ${data.is_hidden ? 'hidden' : 'visible'}.`)
        },
      })
    }
  }

  const menuItems: ActionMenuItem[] = [
    {
      label: channel.isHidden ? 'Unhide Channel' : 'Hide Channel',
      icon: channel.isHidden ? <LuEye size={16} /> : <LuEyeOff size={16} />,
      onClick: handleHide,
    },
  ]

  return (
    <div
      role="button"
      tabIndex={0}
      onClick={onClick}
      className={`group relative flex w-full flex-col overflow-hidden rounded-sm text-left transition-all duration-[260ms] [transition-timing-function:cubic-bezier(.2,.7,.2,1)] ${
        isActive
          ? 'shadow-pop ring-2 ring-crimson-500'
          : 'shadow-card hover:z-10 hover:-translate-y-0.5 hover:shadow-pop'
      }`}
    >
      <div
        className={`relative flex aspect-[4/3] w-full items-center justify-center p-4 ${bgColor}`}
      >
        {/* Highlight + diagonal texture */}
        <div className="pointer-events-none absolute inset-0 bg-[linear-gradient(180deg,rgba(255,255,255,0.06)_0%,rgba(0,0,0,0.28)_100%)]" />
        <div className="pointer-events-none absolute inset-0 bg-[linear-gradient(135deg,rgba(255,255,255,0.06)_0_1px,transparent_1px_14px)] bg-[size:14px_14px] opacity-20" />

        {/* LIVE indicator */}
        <div className="absolute left-2.5 top-2.5 flex items-center gap-1.5 bg-black/70 px-1.5 py-0.5 text-[9px] font-bold uppercase tracking-[0.24em] text-white backdrop-blur-sm">
          <span className="live-dot h-1.5 w-1.5 rounded-full bg-crimson-500" />
          LIVE
        </div>

        {/* Action Menu aligned top-right */}
        <div className="absolute right-2.5 top-2.5 z-30">
          <ActionMenu items={menuItems} />
        </div>

        {/* PLAYING badge */}
        {isActive && (
          <div className="absolute right-2.5 bottom-2.5 bg-crimson-500 px-1.5 py-0.5 text-[9px] font-bold uppercase tracking-[0.24em] text-white">
            Playing
          </div>
        )}

        {/* Logo / wordmark */}
        {channel.logo && !imageError ? (
          <img
            src={channel.logo}
            alt={channel.name}
            className="relative z-10 max-h-16 w-auto max-w-[82%] object-contain drop-shadow-[0_12px_24px_rgba(0,0,0,0.55)] sm:max-h-20"
            onError={() => setImageError(true)}
            loading="lazy"
          />
        ) : (
          <p className="relative z-10 line-clamp-3 px-1 text-center text-base font-black leading-[1.05] tracking-[-0.02em] text-white/92 sm:text-lg">
            {cleanName}
          </p>
        )}

        {/* Hover play overlay (not shown on active — already playing) */}
        {!isActive && (
          <div className="absolute inset-0 z-20 flex items-center justify-center bg-black/45 opacity-0 transition-opacity group-hover:opacity-100">
            <span className="flex h-11 w-11 items-center justify-center rounded-full bg-white text-black shadow-pop">
              <LuPlay size={16} className="ml-0.5" fill="currentColor" />
            </span>
          </div>
        )}
      </div>

      <div className="flex flex-col justify-center bg-ink-50 px-3 py-2.5">
        <p className="line-clamp-1 text-sm font-bold leading-tight tracking-[-0.01em] text-white">
          {channel.name}
        </p>
        <p className="mt-0.5 line-clamp-1 text-[11px] font-medium text-fog-500 sm:text-xs">
          {channel.groupTitle || 'Live channel'}
        </p>
      </div>
    </div>
  )
}
