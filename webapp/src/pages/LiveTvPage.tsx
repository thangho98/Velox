import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router'
import { useLiveTVChannels, useLiveTVEpg, useLiveTVGroups } from '@/hooks/stores/useLiveTV'
import { usePreferences, useUpdatePreferences } from '@/hooks/stores/useAuth'
import type { LiveChannel } from '@/hooks/stores/useLiveTV'
import { CategoryPills } from '@/components/livetv/CategoryPills'
import { ChannelCard } from '@/components/livetv/ChannelCard'
import { EpgTimeline } from '@/components/livetv/EpgTimeline'
import { LiveTVPlayer } from '@/components/livetv/LiveTVPlayer'

export default function LiveTvPage() {
  const navigate = useNavigate()
  const { data: preferences } = usePreferences()
  const { mutate: updatePrefs } = useUpdatePreferences()
  const [selectedGroup, setSelectedGroup] = useState<string>('all')
  const { data: groups = [] } = useLiveTVGroups()

  const queryGroup = selectedGroup === 'all' ? null : selectedGroup
  const { data: channels = [], isLoading: channelsLoading } = useLiveTVChannels(queryGroup)

  const [playingChannelInfo, setPlayingChannelInfo] = useState<LiveChannel | null>(null)

  const { data: epgPrograms = [] } = useLiveTVEpg(playingChannelInfo?.id)

  const displayGroups = useMemo(() => [...groups, 'Hidden'], [groups])

  // Seed first channel once; try to use the last watched channel if it exists and is active.
  useEffect(() => {
    if (!playingChannelInfo && channels.length > 0) {
      let initial = channels.find((c) => c.isActive) || channels[0] // fallback to first active or first

      if (preferences?.last_live_channel_id) {
        const found = channels.find((c) => c.id === preferences.last_live_channel_id && c.isActive)
        if (found) {
          initial = found
        }
      }
      queueMicrotask(() => setPlayingChannelInfo(initial))
    }
  }, [channels, playingChannelInfo, preferences?.last_live_channel_id])

  const currentProgram = useMemo(() => {
    const now = Date.now()
    return (
      epgPrograms.find((p) => {
        const start = new Date(p.start_time).getTime()
        const end = new Date(p.end_time).getTime()
        return start <= now && now < end
      }) ?? null
    )
  }, [epgPrograms])

  const handleChannelClick = (channel: LiveChannel) => {
    updatePrefs({ last_live_channel_id: channel.id })
    navigate(`/livetv/watch/${channel.id}`)
  }

  return (
    <div className="flex flex-col gap-8 pb-8 sm:gap-10">
      {/* HERO PLAYER — break out of Layout padding for cinematic edge-to-edge */}
      <section className="-mx-4 sm:-mx-6 lg:-mx-10">
        <LiveTVPlayer
          channel={playingChannelInfo}
          currentProgram={currentProgram}
          onWatch={(ch, autoFullscreen) => {
            updatePrefs({ last_live_channel_id: ch.id })
            navigate(`/livetv/watch/${ch.id}`, { state: { autoFullscreen } })
          }}
          className="aspect-video max-h-[min(72vh,720px)] w-full lg:aspect-[21/9]"
        />
      </section>

      {/* EPG TIMELINE */}
      <section>
        <EpgTimeline programs={epgPrograms} channelName={playingChannelInfo?.name} />
      </section>

      {/* BROWSE — sticky category pills + channel grid */}
      <section>
        <div className="sticky top-[72px] z-30 -mx-4 bg-ink-0 py-3 sm:-mx-6 lg:-mx-10">
          <CategoryPills
            groups={displayGroups}
            selectedGroup={selectedGroup}
            onSelectGroup={setSelectedGroup}
            className="px-4 sm:px-6 lg:px-10"
          />
        </div>

        <div className="mt-5 flex items-end justify-between">
          <div>
            <p className="mb-1 text-[10px] font-bold uppercase tracking-[0.32em] text-fog-500">
              Live Channels
            </p>
            <h2 className="text-lg font-bold tracking-[-0.02em] text-white sm:text-xl">
              {selectedGroup === 'all' ? 'All Channels' : selectedGroup}
            </h2>
          </div>
          <p className="text-xs font-medium text-fog-500">
            {channels.length} {channels.length === 1 ? 'channel' : 'channels'}
          </p>
        </div>

        {channelsLoading ? (
          <div className="mt-5 grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
            {Array.from({ length: 12 }).map((_, i) => (
              <div key={i} className="aspect-[4/3] animate-pulse bg-ink-100 shadow-card" />
            ))}
          </div>
        ) : channels.length === 0 ? (
          <div className="mt-5 bg-ink-50 py-12 text-center text-fog-500 shadow-card">
            <p className="text-sm">No channels available in this category.</p>
          </div>
        ) : (
          <div className="mt-5 grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
            {channels.map((channel) => (
              <ChannelCard
                key={channel.id}
                channel={channel}
                isActive={channel.id === playingChannelInfo?.id}
                onClick={() => handleChannelClick(channel)}
              />
            ))}
          </div>
        )}
      </section>
    </div>
  )
}
