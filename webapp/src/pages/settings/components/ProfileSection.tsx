import { useState } from 'react'
import { useProfile, useUpdateProfile } from '@/hooks/stores/useAuth'
import { useTranslation } from '@/hooks/useTranslation'
import { SectionHeader, Field, SaveButton, SuccessMsg, inputClass, inputDisabled } from './shared'

export function ProfileSection() {
  const { t } = useTranslation('settings')
  const { data: profile } = useProfile()
  const { mutate: updateProfile, isPending } = useUpdateProfile()
  const [edited, setEdited] = useState<string | null>(null)
  const displayName = edited ?? profile?.display_name ?? ''
  const setDisplayName = (v: string) => setEdited(v)
  const [success, setSuccess] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setSuccess('')
    updateProfile(
      { display_name: displayName },
      {
        onSuccess: () =>
          setSuccess(t('sections.profile.success') || 'Profile updated successfully'),
      },
    )
  }

  return (
    <div className="max-w-xl">
      <SectionHeader
        title={t('sections.profile.title')}
        description={t('sections.profile.description')}
      />
      <form onSubmit={handleSubmit} className="mt-6 space-y-5">
        {success && <SuccessMsg>{success}</SuccessMsg>}
        <Field label={t('fields.username')}>
          <input type="text" value={profile?.username || ''} disabled className={inputDisabled} />
          <p className="mt-1 text-xs text-gray-500">Username cannot be changed</p>
        </Field>
        <Field label={t('fields.displayName')}>
          <input
            type="text"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            className={inputClass}
          />
        </Field>
        <Field label={t('fields.role')}>
          <span
            className={`inline-block rounded px-3 py-1 text-sm ${
              profile?.is_admin
                ? 'bg-purple-500/20 text-purple-400'
                : 'bg-blue-500/20 text-blue-400'
            }`}
          >
            {profile?.is_admin ? 'Administrator' : 'User'}
          </span>
        </Field>
        <SaveButton isPending={isPending} />
      </form>
    </div>
  )
}
