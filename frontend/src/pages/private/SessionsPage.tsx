import { faArrowLeft, faLaptop, faTrashCan } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { useEffect, useState } from 'react'
import { authApi } from '../../api/auth'
import type { UserSession } from '../../api/auth'
import { AuthenticatedLayout } from '../../components/layout/AuthenticatedLayout'
import { Alert } from '../../components/ui/Alert'
import { formatDate, getBrowserName } from '../../lib/format'
import { navigate } from '../../lib/routes'

export function SessionsPage({ onCurrentRevoked }: { onCurrentRevoked: () => void }) {
  const [sessions, setSessions] = useState<UserSession[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [revoking, setRevoking] = useState<number | null>(null)

  useEffect(() => {
    let active = true
    authApi.sessions()
      .then(result => { if (active) setSessions(result.sessions) })
      .catch(reason => { if (active) setError(reason instanceof Error ? reason.message : 'Unable to load sessions.') })
      .finally(() => { if (active) setLoading(false) })
    return () => { active = false }
  }, [])

  async function revoke(session: UserSession) {
    setRevoking(session.id); setError('')
    try {
      await authApi.revokeSession(session.id)
      if (session.current) { onCurrentRevoked(); navigate('/login'); return }
      setSessions(current => current.filter(item => item.id !== session.id))
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : 'Unable to revoke this session.')
    } finally { setRevoking(null) }
  }

  return <AuthenticatedLayout title="Sessions" route="/profile/sessions"><button className="btn btn-ghost btn-sm mb-3 -ml-2" onClick={() => navigate('/profile/edit')}><FontAwesomeIcon icon={faArrowLeft}/>Edit profile</button><section className="card border border-base-300 bg-base-100 shadow-sm"><div className="card-body p-5"><div><h2 className="card-title">Active sessions</h2><p className="mt-1 text-xs leading-5 text-base-content/50">Devices currently signed into your account. Revoke anything you don’t recognize.</p></div>{error && <Alert type="error">{error}</Alert>}{loading ? <div className="grid place-items-center py-12"><span className="loading loading-spinner text-primary"/></div> : sessions.length === 0 ? <p className="py-10 text-center text-sm text-base-content/45">No active sessions found.</p> : <div className="flex flex-col">{sessions.map((session, index) => <div key={session.id} className={`flex gap-3 py-4 ${index ? 'border-t border-base-300' : ''}`}><span className="grid size-10 flex-none place-items-center rounded-full bg-base-200 text-base-content/55"><FontAwesomeIcon icon={faLaptop}/></span><div className="min-w-0 flex-1"><div className="flex items-center gap-2"><strong className="truncate text-sm">{getBrowserName(session.userAgent)}</strong>{session.current && <span className="badge badge-success badge-xs">Current</span>}</div><p className="mt-1 text-[.7rem] text-base-content/45">{session.ipAddress || 'IP unavailable'} · Last active {formatDate(session.lastSeenAt)}</p><p className="mt-1 text-[.65rem] capitalize text-base-content/35">{session.authMethod} · {session.riskLevel} risk</p></div><button className="btn btn-ghost btn-square btn-sm text-error" aria-label={`Revoke ${getBrowserName(session.userAgent)} session`} disabled={revoking === session.id} onClick={() => void revoke(session)}>{revoking === session.id ? <span className="loading loading-spinner loading-xs"/> : <FontAwesomeIcon icon={faTrashCan}/>}</button></div>)}</div>}</div></section></AuthenticatedLayout>
}
