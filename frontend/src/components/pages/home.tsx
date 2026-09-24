import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import {
  faCircleCheck,
  faHouse,
  faMagnifyingGlass,
  faPaperPlane,
  faSquareCaretRight,
} from '@fortawesome/free-solid-svg-icons'
import { useEffect, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import { getSession, login, logout } from '../../api/auth'
import type { Session } from '../../api/auth'

type LoginFormProps = {
  onSuccess: () => void
}

function LoginForm({ onSuccess }: LoginFormProps) {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError('')
    setSubmitting(true)

    try {
      await login({ email, password })
      onSuccess()
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Unable to log in')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="modal-box">
      <h2 className="text-lg font-bold">Log in with email</h2>
      <p className="mt-1 text-left text-sm opacity-60">
        Enter your account details to continue.
      </p>

      <fieldset className="fieldset mt-4" disabled={submitting}>
        <label className="label" htmlFor="login-email">Email</label>
        <input
          id="login-email"
          type="email"
          className="input w-full"
          autoComplete="email"
          required
          value={email}
          onChange={(event) => setEmail(event.target.value)}
        />

        <label className="label" htmlFor="login-password">Password</label>
        <input
          id="login-password"
          type="password"
          className="input w-full"
          autoComplete="current-password"
          required
          value={password}
          onChange={(event) => setPassword(event.target.value)}
        />

        {error && <p role="alert" className="mt-2 text-left text-sm text-error">{error}</p>}

        <button type="submit" className="btn btn-neutral mt-4">
          {submitting ? 'Logging in…' : 'Log in'}
        </button>
      </fieldset>
    </form>
  )
}

function Home() {
  const loginDialog = useRef<HTMLDialogElement>(null)
  const [session, setSession] = useState<Session | null>(null)
  const [checkingSession, setCheckingSession] = useState(true)
  const [sessionError, setSessionError] = useState('')
  const [loggingOut, setLoggingOut] = useState(false)

  async function refreshSession(signal?: AbortSignal) {
    try {
      setSession(await getSession(signal))
    } catch (error) {
      if (error instanceof DOMException && error.name === 'AbortError') return
      setSessionError(error instanceof Error ? error.message : 'Unable to check your session')
    } finally {
      setCheckingSession(false)
    }
  }

  async function handleLogout() {
    setSessionError('')
    setLoggingOut(true)

    try {
      await logout()
      setSession(null)
    } catch (error) {
      setSessionError(error instanceof Error ? error.message : 'Unable to log out')
    } finally {
      setLoggingOut(false)
    }
  }

  useEffect(() => {
    const controller = new AbortController()

    async function checkInitialSession() {
      try {
        setSession(await getSession(controller.signal))
      } catch (error) {
        if (error instanceof DOMException && error.name === 'AbortError') return
        setSessionError(error instanceof Error ? error.message : 'Unable to check your session')
      } finally {
        setCheckingSession(false)
      }
    }

    void checkInitialSession()
    return () => controller.abort()
  }, [])

  return (
    <>
      <nav className="fixed inset-x-0 bottom-4 z-10 flex justify-center px-4" aria-label="Primary navigation">
        <ul className="menu menu-horizontal w-full max-w-xl justify-evenly rounded-full bg-base-200 shadow-lg">
          {[faHouse, faSquareCaretRight, faPaperPlane, faMagnifyingGlass].map((icon) => (
            <li key={icon.iconName}>
              <button type="button" aria-label={icon.iconName}>
                <FontAwesomeIcon className="w-5" icon={icon} />
              </button>
            </li>
          ))}
          <li>
            <button type="button" aria-label="Profile" className="avatar avatar-placeholder">
              <span className="flex w-8 items-center justify-center rounded-full bg-neutral text-xs text-neutral-content">UI</span>
            </button>
          </li>
        </ul>
      </nav>

      <main className="grid min-h-screen grid-cols-1 pb-24 md:grid-cols-2">
        <section className="flex items-center justify-center bg-base-200 p-10">
          <img src="/logo.svg" alt="0xSocial" className="w-64 max-w-[70%]" />
        </section>

        <section className="flex items-center justify-center p-6">
          <div className="flex w-full max-w-sm flex-col gap-5">
            <div>
              <p className="text-sm font-medium uppercase tracking-[0.2em] opacity-50">0xSocial</p>
              <h1 className="mt-2 text-3xl font-bold">Welcome back</h1>
            </div>

            <div className="card border border-base-300 bg-base-100 shadow-sm">
              <div className="card-body gap-2 p-5">
                <div className="flex items-center gap-3">
                  <span
                    className={`status ${session ? 'status-success' : checkingSession ? 'status-warning' : 'status-neutral'}`}
                    aria-hidden="true"
                  />
                  <h2 className="card-title text-base">
                    {checkingSession ? 'Checking session…' : session ? 'Session active' : 'No active session'}
                  </h2>
                </div>

                {session && (
                  <>
                    <p className="truncate text-left text-sm opacity-70">{session.email}</p>
                    <div className="flex items-center justify-between gap-3">
                      <p className="flex items-center gap-2 text-left text-xs text-success">
                        <FontAwesomeIcon icon={faCircleCheck} /> Authenticated securely
                      </p>
                      <button
                        type="button"
                        className="btn btn-ghost btn-sm"
                        disabled={loggingOut}
                        onClick={() => void handleLogout()}
                      >
                        {loggingOut ? 'Logging out…' : 'Log out'}
                      </button>
                    </div>
                  </>
                )}

                {!checkingSession && !session && !sessionError && (
                  <p className="text-left text-sm opacity-60">Log in to start a new session.</p>
                )}

                {sessionError && <p role="alert" className="text-left text-sm text-error">{sessionError}</p>}
              </div>
            </div>

            {!checkingSession && !session && (
              <button
                type="button"
                className="btn btn-neutral w-full"
                onClick={() => loginDialog.current?.showModal()}
              >
                Log in with email
              </button>
            )}

            {session && (
              <p className="text-sm opacity-60">
                Your browser will keep this session until it expires or is revoked.
              </p>
            )}
          </div>
        </section>
      </main>

      <dialog ref={loginDialog} className="modal">
        <LoginForm
          onSuccess={() => {
            loginDialog.current?.close()
            setSessionError('')
            setCheckingSession(true)
            void refreshSession()
          }}
        />
        <form method="dialog" className="modal-backdrop">
          <button type="submit">Close</button>
        </form>
      </dialog>
    </>
  )
}

export default Home
