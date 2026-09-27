import { faArrowRight } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { useState } from 'react'
import type { FormEvent } from 'react'
import { ApiError, authApi } from '../../api/auth'
import { PublicLayout } from '../../components/layout/PublicLayout'
import { Alert } from '../../components/ui/Alert'
import { FormCard } from '../../components/ui/FormCard'
import { PasswordField } from '../../components/ui/PasswordField'
import { useCountdown } from '../../hooks/useCountdown'
import { getQueryParam, navigate, navigatePath } from '../../lib/routes'

export function LoginPage({ onAuthenticated }: { onAuthenticated: () => Promise<void> }) {
  const [email, setEmail] = useState(() => getQueryParam('email'))
  const [password, setPassword] = useState('')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const [status, setStatus] = useState(0)
  const [cooldown, setCooldown] = useCountdown()

  async function submit(event: FormEvent) {
    event.preventDefault(); setBusy(true); setError(''); setStatus(0)
    try {
      await authApi.login(email, password)
      await onAuthenticated()
      const returnTo = getQueryParam('returnTo')
      if (/^\/(?!\/)/.test(returnTo)) navigatePath(returnTo)
      else navigate('/feed')
    } catch (reason) {
      const apiError = reason as ApiError
      setStatus(apiError.status ?? 0)
      setError(apiError.message ?? 'Unable to log in.')
      if (apiError.status === 429) setCooldown(apiError.retryAfter ?? 30)
    } finally { setBusy(false) }
  }

  return <PublicLayout back="/"><FormCard eyebrow="Welcome back" title="Log in to 0xSocial" description="Enter your account details to continue." footer={<>New to 0xSocial? <button className="link link-primary font-semibold" onClick={() => navigate('/register')}>Create an account</button></>}><form className="flex flex-col gap-3" onSubmit={submit}><fieldset className="fieldset"><legend className="fieldset-legend">Email address</legend><input className="input w-full" type="email" autoComplete="email" required value={email} onChange={event => setEmail(event.target.value)} placeholder="you@example.com"/></fieldset><PasswordField password={password} onChange={setPassword} autoComplete="current-password" labelAction={<button className="link link-primary text-xs" type="button" onClick={() => navigate('/forgot-password', { email })}>Forgot password?</button>}/>{error && <Alert type="error">{error}{status === 403 && <> <button type="button" className="link font-semibold" onClick={() => navigate('/check-email', { email })}>Resend verification email</button></>}</Alert>}<button className="btn btn-primary mt-1" type="submit" disabled={busy || cooldown > 0}>{busy ? <><span className="loading loading-spinner loading-sm"/>Logging in…</> : cooldown > 0 ? `Try again in ${cooldown}s` : <>Log in<FontAwesomeIcon icon={faArrowRight}/></>}</button></form></FormCard></PublicLayout>
}
