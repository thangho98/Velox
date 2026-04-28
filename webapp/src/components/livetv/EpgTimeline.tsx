import type { LiveProgram } from '@/hooks/stores/useLiveTV'

interface EpgTimelineProps {
  programs: LiveProgram[]
  channelName?: string | null
}

function formatTime(iso: string) {
  return new Date(iso).toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  })
}

function calcProgress(program: LiveProgram): number {
  const start = new Date(program.start_time).getTime()
  const end = new Date(program.end_time).getTime()
  const now = Date.now()
  if (now < start) return 0
  if (now >= end) return 100
  return ((now - start) / (end - start)) * 100
}

export function EpgTimeline({ programs, channelName }: EpgTimelineProps) {
  if (!programs || programs.length === 0) {
    return null
  }

  const now = Date.now()
  const currentId = programs.find((p) => {
    const s = new Date(p.start_time).getTime()
    const e = new Date(p.end_time).getTime()
    return s <= now && now < e
  })?.id

  return (
    <div>
      <div className="mb-3 flex items-end justify-between">
        <p className="text-[10px] font-bold uppercase tracking-[0.32em] text-fog-500">
          Up Next{channelName ? ` on ${channelName}` : ''}
        </p>
        {programs.length > 3 && (
          <p className="hidden text-[10px] uppercase tracking-[0.28em] text-fog-500 md:block">
            Swipe for more
          </p>
        )}
      </div>

      <div className="hide-scrollbar flex gap-3 overflow-x-auto pb-2">
        {programs.map((program) => {
          const isCurrent = program.id === currentId
          const progress = isCurrent ? calcProgress(program) : 0
          return (
            <article
              key={program.id}
              className={`relative flex min-w-[260px] flex-1 flex-col justify-between overflow-hidden bg-ink-50 p-4 shadow-card transition-colors sm:min-w-[280px] ${
                isCurrent ? 'ring-1 ring-crimson-500' : ''
              }`}
            >
              <div>
                <div className="mb-2 flex items-center justify-between gap-2">
                  <p
                    className={`text-[11px] font-bold uppercase tracking-[0.24em] ${
                      isCurrent ? 'text-crimson-500' : 'text-fog-600'
                    }`}
                  >
                    {isCurrent ? (
                      <span className="inline-flex items-center gap-1.5">
                        <span className="live-dot h-1.5 w-1.5 rounded-full bg-crimson-500" />
                        ON NOW
                      </span>
                    ) : (
                      formatTime(program.start_time)
                    )}
                  </p>
                  <p className="text-[11px] uppercase tracking-[0.18em] text-fog-500">
                    {formatTime(program.start_time)} – {formatTime(program.end_time)}
                  </p>
                </div>
                <p className="line-clamp-2 text-sm font-semibold text-white">{program.title}</p>
                {program.description && (
                  <p className="mt-1 line-clamp-2 text-xs text-fog-500">{program.description}</p>
                )}
              </div>
              {isCurrent && (
                <div className="mt-3 h-1 w-full overflow-hidden bg-ink-300">
                  <div
                    className="h-full bg-crimson-500 transition-all"
                    style={{ width: `${progress}%` }}
                  />
                </div>
              )}
            </article>
          )
        })}
      </div>
    </div>
  )
}
