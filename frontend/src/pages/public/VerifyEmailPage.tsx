import { faArrowRight, faCircleCheck } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { useCallback, useEffect, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import { authApi } from '../../api/auth'
import { PublicLayout } from '../../components/layout/PublicLayout'
import { Alert } from '../../components/ui/Alert'
import { FormCard } from '../../components/ui/FormCard'
import { RegistrationProgress } from '../../components/ui/RegistrationProgress'
import { getQueryParam, navigate } from '../../lib/routes'

type VerificationState = 'entry' | 'verifying' | 'success' | 'error'

export function VerifyEmailPage({ onAuthenticated }: { onAuthenticated: () => Promise<void> }) {
  const [initialToken] = useState(() => getQueryParam('token'))
  const [token, setToken] = useState(initialToken)
  const [state, setState] = useState<VerificationState>(initialToken ? 'verifying' : 'entry')
  const [error, setError] = useState('')
  const attempted = useRef(false)

  const verify = useCallback(async (value: string) => {
    if (!value) return
    setState('verifying'); setError('')
    try { await authApi.verifyEmail(value); setState('success') }
    catch (reason) {
      setError(reason instanceof Error ? reason.message : 'Unable to verify your email.')
      setState('error')
    }
  }, [])

  useEffect(() => {
    if (!initialToken || attempted.current) return
    attempted.current = true
    void verify(initialToken)
  }, [initialToken, verify])

  function submit(event: FormEvent) { event.preventDefault(); void verify(token) }
  async function continueToFeed() { await onAuthenticated(); navigate('/feed') }

  const title = state === 'success' ? 'Your account is ready' : state === 'verifying' ? 'Verifying your email…' : state === 'error' ? 'We couldn’t verify that link' : 'Enter verification code'
  const description = state === 'success' ? 'Welcome to 0xSocial.' : state === 'error' ? error : state === 'verifying' ? 'This will only take a moment.' : 'Paste the verification token from your email.'

  return <PublicLayout back="/login"><RegistrationProgress active={state === 'success' ? 3 : 2}/><FormCard eyebrow="Email verification" title={title} description={description}>{state === 'verifying' && <div className="grid place-items-center py-8"><span className="loading loading-spinner loading-lg text-primary"/></div>}{state === 'success' && <><div className="mx-auto grid size-20 place-items-center rounded-full bg-success/10 text-3xl text-success"><FontAwesomeIcon icon={faCircleCheck}/></div><button className="btn btn-primary" onClick={() => void continueToFeed()}>Go to your feed<FontAwesomeIcon icon={faArrowRight}/></button></>}{state === 'error' && <><Alert type="error">{error}</Alert><button className="btn btn-primary" onClick={() => navigate('/check-email')}>Request another email</button><button className="btn btn-ghost" onClick={() => navigate('/login')}>Back to login</button></>}{state === 'entry' && <form className="flex flex-col gap-4" onSubmit={submit}><fieldset className="fieldset"><legend className="fieldset-legend">Verification token</legend><input className="input w-full font-mono" required value={token} onChange={event => setToken(event.target.value)} placeholder="Paste token here"/></fieldset><button className="btn btn-primary" type="submit">Verify email<FontAwesomeIcon icon={faArrowRight}/></button></form>}</FormCard></PublicLayout>
}
