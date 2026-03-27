import { useTranslation as useI18nTranslation } from 'react-i18next'

// Define namespaces for type safety
export type Namespace =
  | 'common'
  | 'auth'
  | 'navigation'
  | 'settings'
  | 'media'
  | 'watch'
  | 'errors'
  | 'wizard'

export function useTranslation(ns: Namespace = 'common') {
  return useI18nTranslation(ns)
}
