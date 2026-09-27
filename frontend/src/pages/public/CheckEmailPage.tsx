import { faEnvelope, faRotateRight } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { useState } from 'react'
import { authApi } from '../../api/auth'
import { PublicLayout } from '../../components/layout/PublicLayout'
import { Alert } from '../../components/ui/Alert'
import { FormCard } from '../../components/ui/FormCard'
import { RegistrationProgress } from '../../components/ui/RegistrationProgress'
import { useCountdown } from '../../hooks/useCountdown'
import { getQueryParam, navigate } from '../../lib/routes'

export function CheckEmailPage() {
  const email = getQueryParam('email')
  const [countdown, setCountdown] = useCountdown()
  const [message, setMessage] = useState('')
  const [busy, setBusy] = useState(false)

  async function resend() {
    if (!email || countdown) return
    setBusy(true); setMessage('')
    try {
      await authApi.resendVerification(email)
      setCountdown(60)
      setMessage('A new verification email is on its way.')
    } catch (reason) {
      setMessage(reason instanceof Error ? reason.message : 'Unable to resend the email.')
    } finally { setBusy(false) }
  }

  return <PublicLayout back="/register"><RegistrationProgress active={2}/><FormCard eyebrow="Verify your email" title="Check your email" description={<>We sent a verification link to <strong className="text-base-content">{email || 'your email address'}</strong>.</>}><div className="mx-auto grid size-20 place-items-center rounded-full bg-primary/10 text-3xl text-primary"><FontAwesomeIcon icon={faEnvelope}/></div>{message && <Alert type={message.includes('way') ? 'success' : 'error'}>{message}</Alert>}<a className="btn btn-primary" href="mailto:"><FontAwesomeIcon icon={faEnvelope}/>Open email app</a><button className="btn btn-outline" onClick={() => navigate('/verify-email')}>Enter verification code</button><button className="btn btn-ghost btn-sm" disabled={busy || countdown > 0} onClick={() => void resend()}>{busy ? <span className="loading loading-spinner loading-xs"/> : <FontAwesomeIcon icon={faRotateRight}/>} {countdown > 0 ? `Resend in ${countdown}s` : 'Resend email'}</button><button className="link link-hover text-xs text-base-content/50" onClick={() => navigate('/register')}>Use a different email</button></FormCard></PublicLayout>
}
