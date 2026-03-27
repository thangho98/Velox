import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import LanguageDetector from 'i18next-browser-languagedetector'

// English translations
import enCommon from '../locales/en/common.json'
import enAuth from '../locales/en/auth.json'
import enNavigation from '../locales/en/navigation.json'
import enSettings from '../locales/en/settings.json'
import enMedia from '../locales/en/media.json'
import enWatch from '../locales/en/watch.json'
import enErrors from '../locales/en/errors.json'
import enWizard from '../locales/en/wizard.json'

// Vietnamese translations
import viCommon from '../locales/vi/common.json'
import viAuth from '../locales/vi/auth.json'
import viNavigation from '../locales/vi/navigation.json'
import viSettings from '../locales/vi/settings.json'
import viMedia from '../locales/vi/media.json'
import viWatch from '../locales/vi/watch.json'
import viErrors from '../locales/vi/errors.json'
import viWizard from '../locales/vi/wizard.json'

const resources = {
  en: {
    common: enCommon,
    auth: enAuth,
    navigation: enNavigation,
    settings: enSettings,
    media: enMedia,
    watch: enWatch,
    errors: enErrors,
    wizard: enWizard,
  },
  vi: {
    common: viCommon,
    auth: viAuth,
    navigation: viNavigation,
    settings: viSettings,
    media: viMedia,
    watch: viWatch,
    errors: viErrors,
    wizard: viWizard,
  },
}

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    fallbackLng: 'en',
    defaultNS: 'common',
    detection: {
      order: ['localStorage', 'navigator', 'htmlTag'],
      lookupLocalStorage: 'velox-language',
      caches: ['localStorage'],
    },
    interpolation: {
      escapeValue: false, // React already escapes
    },
  })

export default i18n
