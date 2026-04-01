import { useState } from 'react'
import { useChangePassword } from '@/hooks/stores/useAuth'
import { useTranslation } from '@/hooks/useTranslation'
import { SectionHeader, Field, SaveButton, SuccessMsg, ErrorMsg, inputClass } from './shared'

export function SecuritySection() {
  const { t } = useTranslation('settings')
  const { mutate: changePassword, isPending } = useChangePassword()
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setSuccess('')
    if (newPassword !== confirmPassword) {
      setError('Passwords do not match')
      return
    }
    if (newPassword.length < 8) {
      setError('Password must be at least 8 characters')
      return
    }
    changePassword(
      { old_password: oldPassword, new_password: newPassword },
      {
        onSuccess: () => {
          setSuccess('Password changed successfully')
          setOldPassword('')
          setNewPassword('')
          setConfirmPassword('')
        },
        onError: (err: Error) => setError(err.message),
      },
    )
  }

  return (
    <div className="max-w-xl">
      <SectionHeader
        title={t('sections.security.title')}
        description={t('sections.security.description')}
      />
      <form onSubmit={handleSubmit} className="mt-6 space-y-5">
        {error && <ErrorMsg>{error}</ErrorMsg>}
        {success && <SuccessMsg>{success}</SuccessMsg>}
        <Field label={t('fields.currentPassword')}>
          <input
            type="password"
            value={oldPassword}
            onChange={(e) => setOldPassword(e.target.value)}
            className={inputClass}
            required
          />
        </Field>
        <Field label={t('fields.newPassword')}>
          <input
            type="password"
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            className={inputClass}
            required
            minLength={8}
          />
        </Field>
        <Field label={t('fields.confirmPassword')}>
          <input
            type="password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            className={inputClass}
            required
          />
        </Field>
        <SaveButton isPending={isPending} label="Change Password" />
      </form>
    </div>
  )
}
