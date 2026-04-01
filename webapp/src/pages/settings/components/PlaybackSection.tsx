import { useState } from 'react'
import { LuCheck } from 'react-icons/lu'
import { usePlaybackSettings, useUpdatePlaybackSettings } from '@/hooks/stores/useSettings'
import { useTranslation } from '@/hooks/useTranslation'
import { SectionHeader, Spinner } from './shared'

export function PlaybackSection() {
  const { t } = useTranslation('settings')
  const { data: settings, isLoading } = usePlaybackSettings()
  const { mutate: updateSettings, isPending: isSaving } = useUpdatePlaybackSettings()
  const [saved, setSaved] = useState(false)

  const handleChange = (mode: 'auto' | 'direct_play') => {
    updateSettings(
      { playback_mode: mode },
      {
        onSuccess: () => {
          setSaved(true)
          setTimeout(() => setSaved(false), 2000)
        },
      },
    )
  }

  if (isLoading) return <Spinner />

  const current = settings?.playback_mode || 'auto'

  return (
    <div className="max-w-2xl space-y-6">
      <SectionHeader
        title={t('sections.playback.title')}
        description={t('sections.playback.description')}
      />

      <div className="rounded-lg bg-netflix-dark p-6">
        <h3 className="mb-1 text-base font-semibold text-white">{t('fields.playbackMode')}</h3>
        <p className="mb-4 text-sm text-gray-400">{t('playback.policyDescription')}</p>

        <div className="space-y-3">
          {[
            {
              value: 'auto' as const,
              label: t('playback.auto'),
              description: t('playback.auto'),
            },
            {
              value: 'direct_play' as const,
              label: t('playback.directPlay'),
              description: t('playback.directPlay'),
            },
          ].map((option) => (
            <button
              key={option.value}
              onClick={() => handleChange(option.value)}
              disabled={isSaving}
              className={`flex w-full items-start gap-3 rounded-lg border-2 p-4 text-left transition-colors ${
                current === option.value
                  ? 'border-netflix-red bg-netflix-red/10'
                  : 'border-netflix-gray/50 bg-netflix-black/30 hover:border-white/20'
              }`}
            >
              <div
                className={`mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded-full border-2 ${
                  current === option.value ? 'border-netflix-red bg-netflix-red' : 'border-gray-500'
                }`}
              >
                {current === option.value && <div className="h-2 w-2 rounded-full bg-white" />}
              </div>
              <div>
                <p className="text-sm font-medium text-white">{option.label}</p>
                <p className="mt-0.5 text-xs text-gray-400">{option.description}</p>
              </div>
            </button>
          ))}
        </div>

        {saved && (
          <div className="mt-3 flex items-center gap-1.5 text-sm text-green-400">
            <LuCheck size={16} /> {t('actions.saved')}
          </div>
        )}
      </div>
    </div>
  )
}
