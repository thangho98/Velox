import { useState, useRef } from 'react'
import {
  useCinemaSettings,
  useUpdateCinemaSettings,
  useUploadCinemaIntro,
} from '@/hooks/stores/useSettings'
import { Toggle } from '@/components/ui/Toggle'
import { Select } from '@/components/ui/Select'
import { useTranslation } from '@/hooks/useTranslation'
import { SectionHeader, Spinner } from './shared'

export function CinemaSection() {
  const { t } = useTranslation('settings')
  const { data: settings, isLoading } = useCinemaSettings()
  const { mutate: updateSettings } = useUpdateCinemaSettings()
  const { mutate: uploadIntro, isPending: isUploading } = useUploadCinemaIntro()
  const fileRef = useRef<HTMLInputElement>(null)
  const [saved, setSaved] = useState(false)

  if (isLoading) return <Spinner />

  return (
    <div className="max-w-2xl space-y-6">
      <SectionHeader
        title={t('sections.cinema.title')}
        description={t('sections.cinema.description')}
      />

      {/* Enable toggle */}
      <div className="rounded-lg bg-netflix-dark p-6">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-base font-semibold text-white">{t('cinema.enabled')}</h3>
            <p className="mt-1 text-sm text-gray-400">{t('cinema.enabledDescription')}</p>
          </div>
          <Toggle
            enabled={settings?.enabled ?? false}
            onChange={(v) => {
              updateSettings(
                { enabled: v },
                {
                  onSuccess: () => {
                    setSaved(true)
                    setTimeout(() => setSaved(false), 2000)
                  },
                },
              )
            }}
          />
        </div>
        {saved && <p className="mt-2 text-sm text-green-400">{t('actions.saved')}</p>}
      </div>

      {/* Max trailers */}
      <div className="rounded-lg bg-netflix-dark p-6">
        <h3 className="mb-1 text-base font-semibold text-white">{t('cinema.maxTrailers')}</h3>
        <p className="mb-3 text-sm text-gray-400">{t('cinema.maxTrailersDescription')}</p>
        <Select
          value={settings?.max_trailers ?? '2'}
          onChange={(e) => updateSettings({ max_trailers: e.target.value })}
          className="rounded-lg bg-[#2a2a2a] px-4 py-2 text-white outline-none"
        >
          <option value="0">{t('options.quality.original')}</option>
          <option value="1">1</option>
          <option value="2">2</option>
          <option value="3">3</option>
        </Select>
      </div>

      {/* Custom intro video */}
      <div className="rounded-lg bg-netflix-dark p-6">
        <h3 className="mb-1 text-base font-semibold text-white">{t('cinema.customIntro')}</h3>
        <p className="mb-3 text-sm text-gray-400">{t('cinema.customIntroDescription')}</p>
        <div className="flex items-center gap-4">
          <input
            ref={fileRef}
            type="file"
            accept="video/mp4,video/webm,video/quicktime"
            className="hidden"
            onChange={(e) => {
              const file = e.target.files?.[0]
              if (file) uploadIntro(file)
            }}
          />
          <button
            onClick={() => fileRef.current?.click()}
            disabled={isUploading}
            className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-500 disabled:opacity-50"
          >
            {isUploading
              ? t('actions.uploading')
              : settings?.has_intro
                ? t('actions.upload')
                : t('cinema.uploadIntro')}
          </button>
          {settings?.has_intro && (
            <span className="text-sm text-green-400">{t('cinema.hasIntro')}</span>
          )}
        </div>
      </div>
    </div>
  )
}
