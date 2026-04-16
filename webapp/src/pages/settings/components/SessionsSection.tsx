import { useSessions, useRevokeSession } from '@/hooks/stores/useAuth'
import { useTranslation } from '@/hooks/useTranslation'
import { LuMonitor } from 'react-icons/lu'
import { SectionHeader, Spinner } from './shared'

export function SessionsSection() {
  const { t } = useTranslation('settings')
  const { data: sessions, isLoading } = useSessions()
  const { mutate: revokeSession } = useRevokeSession()

  return (
    <div className="max-w-2xl">
      <SectionHeader
        title={t('sections.sessions.title')}
        description={t('sections.sessions.description')}
      />
      {isLoading ? (
        <Spinner />
      ) : (
        <div className="mt-6 space-y-3">
          {sessions?.map((session) => (
            <div
              key={session.id}
              className="flex items-center justify-between rounded-lg bg-netflix-dark p-4"
            >
              <div className="flex items-center gap-4">
                <div className="flex h-10 w-10 items-center justify-center rounded bg-netflix-gray">
                  <LuMonitor size={20} className="text-gray-400" />
                </div>
                <div>
                  <p className="text-sm font-medium text-white">
                    {session.device_name || t('sessions.unknownDevice')}
                  </p>
                  <p className="text-xs text-gray-400">{session.ip_address}</p>
                  <p className="text-xs text-gray-500">
                    {t('sessions.lastActive')} {new Date(session.last_active_at).toLocaleString()}
                  </p>
                </div>
              </div>
              <button
                onClick={() => revokeSession(session.id)}
                className="rounded bg-netflix-gray px-3 py-1.5 text-sm text-white transition-colors hover:bg-netflix-red"
              >
                {t('actions.revoke')}
              </button>
            </div>
          ))}
          {sessions?.length === 0 && (
            <p className="py-8 text-center text-sm text-gray-400">{t('sessions.noSessions')}</p>
          )}
        </div>
      )}
    </div>
  )
}
