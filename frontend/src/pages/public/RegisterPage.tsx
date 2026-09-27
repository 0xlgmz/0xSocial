import type { IconDefinition } from '@fortawesome/fontawesome-svg-core'
import {
  faArrowRight, faAt, faPlus, faShieldHalved, faUserGroup,
} from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { useState } from 'react'
import type { FormEvent } from 'react'
import { ApiError, authApi } from '../../api/auth'
import { PublicLayout } from '../../components/layout/PublicLayout'
import { Alert } from '../../components/ui/Alert'
import { FormCard } from '../../components/ui/FormCard'
import { PasswordChecklist } from '../../components/ui/PasswordChecklist'
import { PasswordField } from '../../components/ui/PasswordField'
import { RegistrationProgress } from '../../components/ui/RegistrationProgress'
import { normalizeHandle, validateHandle } from '../../lib/handle'
import { validatePassword } from '../../lib/password'
import { navigate } from '../../lib/routes'

function Benefit({ icon, text }: { icon: IconDefinition; text: string }) {
  return (
    <div className="rounded-box bg-base-200 p-3 text-center text-[.6rem] text-base-content/55">
      <FontAwesomeIcon icon={icon} className="mb-2 block w-full text-sm text-primary"/>{text}
    </div>
  )
}

export function RegisterPage() {
  const [email, setEmail] = useState('')
  const [handle, setHandle] = useState('')
  const [password, setPassword] = useState('')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const [handleError, setHandleError] = useState('')

  async function submit(event: FormEvent) {
    event.preventDefault()
    const normalizedHandle = normalizeHandle(handle)
    const handleValidation = validateHandle(normalizedHandle)
    if (handleValidation) {
      setHandleError(handleValidation)
      return
    }

    const passwordValidation = validatePassword(password)
    if (passwordValidation) {
      setError(passwordValidation)
      return
    }

    setBusy(true)
    setError('')
    setHandleError('')
    try {
      await authApi.register({ email, password, handle: normalizedHandle })
      navigate('/check-email', { email })
    } catch (reason) {
      if (reason instanceof ApiError && /handle/i.test(reason.message)) {
        setHandleError(reason.message)
      } else {
        setError(reason instanceof Error ? reason.message : 'Unable to create your account.')
      }
    } finally {
      setBusy(false)
    }
  }

  return (
    <PublicLayout back="/">
      <RegistrationProgress active={1}/>
      <FormCard
        eyebrow="Join 0xSocial"
        title="Create your account"
        description="Start building your profile and connecting with people."
        footer={<>Already have an account? <button className="link link-primary font-semibold" onClick={() => navigate('/login')}>Log in</button></>}
      >
        <div className="hidden grid-cols-3 gap-2 sm:grid">
          <Benefit icon={faUserGroup} text="Follow conversations"/>
          <Benefit icon={faPlus} text="Share updates"/>
          <Benefit icon={faShieldHalved} text="Stay in control"/>
        </div>

        <form className="flex flex-col gap-3" onSubmit={submit}>
          <fieldset className="fieldset">
            <legend className="fieldset-legend">Email address</legend>
            <input className="input w-full" type="email" autoComplete="email" required value={email} onChange={event => setEmail(event.target.value)} placeholder="you@example.com"/>
          </fieldset>

          <fieldset className="fieldset">
            <legend className="fieldset-legend">Handle</legend>
            <label className={`input w-full ${handleError ? 'input-error' : ''}`}>
              <FontAwesomeIcon icon={faAt} className="text-base-content/35"/>
              <input
                className="grow"
                type="text"
                autoCapitalize="none"
                autoCorrect="off"
                spellCheck={false}
                required
                value={handle}
                onChange={event => { setHandle(event.target.value.toLowerCase()); setHandleError('') }}
                placeholder="your_handle"
                aria-invalid={Boolean(handleError)}
                aria-describedby="handle-help"
              />
            </label>
            <p id="handle-help" className={`fieldset-label ${handleError ? 'text-error' : ''}`}>
              {handleError || '3–30 characters · letters, numbers, and underscores'}
            </p>
          </fieldset>

          <PasswordField password={password} onChange={setPassword}/>
          <PasswordChecklist password={password}/>

          {error && <Alert type="error">
            {error}
            {error.includes('already exists') && <> <button className="link font-semibold" type="button" onClick={() => navigate('/login', { email })}>Log in instead</button></>}
          </Alert>}

          <button className="btn btn-primary mt-1" type="submit" disabled={busy}>
            {busy ? <><span className="loading loading-spinner loading-sm"/>Creating account…</> : <>Create account<FontAwesomeIcon icon={faArrowRight}/></>}
          </button>
        </form>
      </FormCard>
    </PublicLayout>
  )
}
