import { useState } from 'react'
import { usePreferences, useUpdatePreferences } from '@/hooks/stores/useAuth'
import { useUIStore } from '@/stores/ui'
import { useTranslation } from '@/hooks/useTranslation'
import { Select } from '@/components/ui/Select'
import { LanguageSwitcher } from '@/components/LanguageSwitcher'
import { SectionHeader, Field, SaveButton } from './shared'

export function PreferencesSection() {
  const { t } = useTranslation('settings')
  const { data: preferences } = usePreferences()
  const { mutate: updatePreferences, isPending } = useUpdatePreferences()
  const { theme, setTheme } = useUIStore()

  type PrefsEdits = {
    subtitle_language?: string
    audio_language?: string
    max_streaming_quality?: string
    theme?: 'light' | 'dark' | 'system'
  }
  const [edits, setEdits] = useState<PrefsEdits>({})
  const prefs = {
    subtitle_language: edits.subtitle_language ?? preferences?.subtitle_language ?? '',
    audio_language: edits.audio_language ?? preferences?.audio_language ?? '',
    max_streaming_quality:
      edits.max_streaming_quality ?? preferences?.max_streaming_quality ?? 'original',
    theme: edits.theme ?? theme,
  }
  const setPrefs = (patch: PrefsEdits) => setEdits((prev) => ({ ...prev, ...patch }))

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    updatePreferences({
      user_id: preferences?.user_id || 0,
      subtitle_language: prefs.subtitle_language,
      audio_language: prefs.audio_language,
      max_streaming_quality: prefs.max_streaming_quality,
      theme: prefs.theme,
      language: preferences?.language || 'en',
    })
    setTheme(prefs.theme)
  }

  return (
    <div className="max-w-xl">
      <SectionHeader
        title={t('sections.preferences.title')}
        description={t('sections.preferences.description')}
      />
      <form onSubmit={handleSubmit} className="mt-6 space-y-5">
        <Field label={t('fields.subtitleLanguage')}>
          <Select
            value={prefs.subtitle_language}
            onChange={(e) => setPrefs({ subtitle_language: e.target.value })}
            className="w-full"
          >
            <option value="">{t('options.language.auto')}</option>
            <option value="vi">{t('options.language.vi')}</option>
            <option value="en">{t('options.language.en')}</option>
          </Select>
        </Field>
        <Field label={t('fields.audioLanguage')}>
          <Select
            value={prefs.audio_language}
            onChange={(e) => setPrefs({ audio_language: e.target.value })}
            className="w-full"
          >
            <option value="">{t('options.language.auto')}</option>
            <option value="vi">{t('options.language.vi')}</option>
            <option value="en">{t('options.language.en')}</option>
          </Select>
        </Field>
        <Field label={t('fields.maxQuality')}>
          <Select
            value={prefs.max_streaming_quality}
            onChange={(e) => setPrefs({ max_streaming_quality: e.target.value })}
            className="w-full"
          >
            <option value="original">{t('options.quality.original')}</option>
            <option value="4k">{t('options.quality.4k')}</option>
            <option value="1080p">{t('options.quality.1080p')}</option>
            <option value="720p">{t('options.quality.720p')}</option>
            <option value="480p">{t('options.quality.480p')}</option>
          </Select>
        </Field>
        <Field label={t('fields.theme')}>
          <Select
            value={prefs.theme}
            onChange={(e) => setPrefs({ theme: e.target.value as 'light' | 'dark' | 'system' })}
            className="w-full"
          >
            <option value="system">{t('options.theme.system')}</option>
            <option value="dark">{t('options.theme.dark')}</option>
            <option value="light">{t('options.theme.light')}</option>
          </Select>
        </Field>
        <Field label={t('fields.language')}>
          <LanguageSwitcher />
        </Field>
        <SaveButton isPending={isPending} />
      </form>
    </div>
  )
}
