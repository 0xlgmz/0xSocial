import { faArrowRight, faCircleCheck, faLock } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { useState } from 'react'
import type { FormEvent } from 'react'
import { authApi } from '../../api/auth'
import { PublicLayout } from '../../components/layout/PublicLayout'
import { Alert } from '../../components/ui/Alert'
import { FormCard } from '../../components/ui/FormCard'
import { PasswordChecklist } from '../../components/ui/PasswordChecklist'
import { PasswordField } from '../../components/ui/PasswordField'
import { validatePassword } from '../../lib/password'
import { getQueryParam, navigate } from '../../lib/routes'

export function ResetPasswordPage() {
  const [token] = useState(() => getQueryParam('token'))
  const [password, setPassword] = useState('')
  const [busy, setBusy] = useState(false)
  const [done, setDone] = useState(false)
  const [error, setError] = useState('')

  async function submit(event: FormEvent) {
    event.preventDefault()
    const validation = validatePassword(password)
    if (validation) { setError(validation); return }
    if (!token) { setError('This reset link is missing its token. Request a new email.'); return }
    setBusy(true); setError('')
    try { await authApi.resetPassword(token, password); setDone(true) }
    catch (reason) { setError(reason instanceof Error ? reason.message : 'Unable to reset your password.') }
    finally { setBusy(false) }
  }

  return <PublicLayout back="/login"><FormCard eyebrow="Account recovery" title={done ? 'Password changed' : 'Choose a new password'} description={done ? 'Your password was updated and all existing sessions were revoked.' : 'Choose a secure password you haven’t used before.'}>{done ? <><div className="mx-auto grid size-20 place-items-center rounded-full bg-success/10 text-3xl text-success"><FontAwesomeIcon icon={faCircleCheck}/></div><button className="btn btn-primary" onClick={() => navigate('/login')}>Continue to login<FontAwesomeIcon icon={faArrowRight}/></button></> : <form className="flex flex-col gap-3" onSubmit={submit}><PasswordField label="New password" password={password} onChange={setPassword}/><PasswordChecklist password={password}/>{error && <Alert type="error">{error}</Alert>}<button className="btn btn-primary" type="submit" disabled={busy}>{busy ? <span className="loading loading-spinner loading-sm"/> : <FontAwesomeIcon icon={faLock}/>}Change password</button></form>}</FormCard></PublicLayout>
}
