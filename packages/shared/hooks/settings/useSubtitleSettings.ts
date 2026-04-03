/**
 * Subtitle provider settings hooks
 */

import { createSettingsHooks } from './factory'

// --- Types ---

interface SubdlSettings {
  api_key: string
  has_builtin: boolean
}

interface DeepLSettings {
  api_key: string
}

interface AutoSubSettings {
  languages: string
}

// --- Hooks ---

export const [useSubdlSettings, useUpdateSubdlSettings] = createSettingsHooks<
  SubdlSettings,
  { api_key: string }
>('subdl')

export const [useDeepLSettings, useUpdateDeepLSettings] =
  createSettingsHooks<DeepLSettings>('deepl')

export const [useAutoSubSettings, useUpdateAutoSubSettings] =
  createSettingsHooks<AutoSubSettings>('auto-subtitles')
