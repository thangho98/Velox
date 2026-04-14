/**
 * Subtitle provider settings hooks
 */

import { createSettingsHooks } from "./factory";

// --- Types ---

interface SubdlSettings {
  api_key: string;
  has_builtin: boolean;
}

interface DeepLSettings {
  api_key: string;
}

interface AITranslationSettings {
  provider:
    | ""
    | "openai_compatible"
    | "gemini_compatible"
    | "anthropic_compatible";
  api_key: string;
  base_url: string;
  model: string;
}

interface AutoSubSettings {
  languages: string;
}

interface AutoTranslateSettings {
  enabled: boolean;
  languages: string;
}

// --- Hooks ---

export const [useSubdlSettings, useUpdateSubdlSettings] = createSettingsHooks<
  SubdlSettings,
  { api_key: string }
>("subdl");

export const [useDeepLSettings, useUpdateDeepLSettings] =
  createSettingsHooks<DeepLSettings>("deepl");

export const [useAITranslationSettings, useUpdateAITranslationSettings] =
  createSettingsHooks<AITranslationSettings>("ai-translation");

export const [useAutoSubSettings, useUpdateAutoSubSettings] =
  createSettingsHooks<AutoSubSettings>("auto-subtitles");

export const [useAutoTranslateSettings, useUpdateAutoTranslateSettings] =
  createSettingsHooks<AutoTranslateSettings>("auto-translate");
