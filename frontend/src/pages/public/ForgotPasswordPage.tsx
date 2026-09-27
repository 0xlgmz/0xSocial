import { faEnvelope } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { useState } from 'react'
import type { FormEvent } from 'react'
import { authApi } from '../../api/auth'
import { PublicLayout } from '../../components/layout/PublicLayout'
import { Alert } from '../../components/ui/Alert'
import { FormCard } from '../../components/ui/FormCard'
import { getQueryParam, navigate } from '../../lib/routes'

export function ForgotPasswordPage() {
  const [email, setEmail] = useState(() => getQueryParam('email'))
  const [busy, setBusy] = useState(false)
  const [sent, setSent] = useState(false)
  const [error, setError] = useState('')

  async function submit(event: FormEvent) {
    event.preventDefault(); setBusy(true); setError('')
    try { await authApi.forgotPassword(email); setSent(true) }
    catch (reason) { setError(reason instanceof Error ? reason.message : 'Unable to send reset instructions.') }
    finally { setBusy(false) }
  }

  return <PublicLayout back="/login"><FormCard eyebrow="Account recovery" title={sent ? 'Check your inbox' : 'Reset your password'} description={sent ? 'If an account matches that email, we’ll send password-reset instructions.' : 'Enter your email and we’ll help you get back into your account.'}>{sent ? <><div className="mx-auto grid size-20 place-items-center rounded-full bg-success/10 text-3xl text-success"><FontAwesomeIcon icon={faEnvelope}/></div><Alert type="info">For your privacy, we show the same confirmation whether or not an account exists.</Alert><button className="btn btn-primary" onClick={() => navigate('/login', { email })}>Back to login</button></> : <form className="flex flex-col gap-4" onSubmit={submit}><fieldset className="fieldset"><legend className="fieldset-legend">Email address</legend><input className="input w-full" type="email" required autoComplete="email" value={email} onChange={event => setEmail(event.target.value)} placeholder="you@example.com"/></fieldset>{error && <Alert type="error">{error}</Alert>}<button className="btn btn-primary" type="submit" disabled={busy}>{busy ? <span className="loading loading-spinner loading-sm"/> : <FontAwesomeIcon icon={faEnvelope}/>}Send reset instructions</button></form>}</FormCard></PublicLayout>
}
