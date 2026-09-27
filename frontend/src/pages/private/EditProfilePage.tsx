import { faArrowLeft, faArrowRight, faCalendarDays, faClock, faEnvelope, faFloppyDisk, faLaptop, faRightFromBracket } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import type { IconDefinition } from '@fortawesome/fontawesome-svg-core'
import { useState } from 'react'
import type { FormEvent } from 'react'
import type { Profile, UpdateProfileInput } from '../../api/auth'
import { AuthenticatedLayout } from '../../components/layout/AuthenticatedLayout'
import { Alert } from '../../components/ui/Alert'
import { ThemeSelector } from '../../components/ui/ThemeSelector'
import { ProfileAvatar } from '../../components/user/ProfileAvatar'
import { formatDate } from '../../lib/format'
import { navigate } from '../../lib/routes'

function characterCount(value: string) { return [...value.trim()].length }

function ProfileDetail({ icon, label, value }: { icon: IconDefinition; label: string; value: string }) {
  return <div className="flex min-h-14 items-center gap-4 px-3 py-2"><span className="grid size-10 flex-none place-items-center rounded-full bg-base-200 text-base-content/55"><FontAwesomeIcon icon={icon}/></span><span className="min-w-0"><strong className="block text-sm">{label}</strong><small className="block truncate text-base-content/45">{value}</small></span></div>
}

type EditProfilePageProps = { profile: Profile; onUpdate: (input: UpdateProfileInput) => Promise<void>; onLogout: () => Promise<void> }

export function EditProfilePage({ profile, onUpdate, onLogout }: EditProfilePageProps) {
  const [displayName, setDisplayName] = useState(profile.displayName)
  const [bio, setBio] = useState(profile.bio)
  const [busy, setBusy] = useState(false)
  const [loggingOut, setLoggingOut] = useState(false)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const displayNameLength = characterCount(displayName)
  const bioLength = characterCount(bio)
  const valid = displayNameLength <= 80 && bioLength <= 500
  const unchanged = displayName === profile.displayName && bio === profile.bio

  async function submit(event: FormEvent) {
    event.preventDefault()
    if (!valid || unchanged) return
    setBusy(true); setError(''); setNotice('')
    try { await onUpdate({ displayName, bio }); setNotice('Your profile has been updated.') }
    catch (reason) { setError(reason instanceof Error ? reason.message : 'Unable to update your profile.') }
    finally { setBusy(false) }
  }

  async function logout() {
    setLoggingOut(true); setError('')
    try { await onLogout() }
    catch (reason) { setError(reason instanceof Error ? reason.message : 'Unable to log out.'); setLoggingOut(false) }
  }

  return (
    <AuthenticatedLayout title="Edit profile" route="/profile/edit">
      <button className="btn btn-ghost btn-sm mb-3 -ml-2" onClick={() => navigate('/profile')}><FontAwesomeIcon icon={faArrowLeft}/>Profile</button>
      <div className="flex flex-col gap-4">
        <section className="card border border-base-300 bg-base-100 shadow-sm">
          <form className="card-body gap-5 p-5 sm:p-7" onSubmit={submit}>
            <div className="flex items-center gap-4"><ProfileAvatar displayName={profile.displayName} fallbackText={profile.handle || profile.email} avatarUrl={profile.avatarUrl} size="sm"/><div><h1 className="text-xl font-semibold">Edit profile</h1><p className="text-xs text-base-content/50">@{profile.handle}</p></div></div>
            <fieldset className="fieldset"><div className="flex justify-between"><legend className="fieldset-legend">Display name</legend><span className={`text-[.65rem] ${displayNameLength > 80 ? 'text-error' : 'text-base-content/40'}`}>{displayNameLength}/80</span></div><input className={`input w-full ${displayNameLength > 80 ? 'input-error' : ''}`} value={displayName} onChange={event => setDisplayName(event.target.value)} autoComplete="name" aria-invalid={displayNameLength > 80} placeholder="Your display name"/></fieldset>
            <fieldset className="fieldset"><div className="flex justify-between"><legend className="fieldset-legend">Bio</legend><span className={`text-[.65rem] ${bioLength > 500 ? 'text-error' : 'text-base-content/40'}`}>{bioLength}/500</span></div><textarea className={`textarea min-h-32 w-full resize-y ${bioLength > 500 ? 'textarea-error' : ''}`} value={bio} onChange={event => setBio(event.target.value)} aria-invalid={bioLength > 500} placeholder="Tell people a little about yourself"/></fieldset>
            {error && <Alert type="error">{error}</Alert>}{notice && <Alert type="success">{notice}</Alert>}
            <div className="flex justify-end"><button className="btn btn-primary" type="submit" disabled={busy || !valid || unchanged}>{busy ? <span className="loading loading-spinner loading-sm"/> : <FontAwesomeIcon icon={faFloppyDisk}/>}Save changes</button></div>
          </form>
        </section>

        <section className="card border border-base-300 bg-base-100 shadow-sm"><div className="card-body gap-0 p-3"><ProfileDetail icon={faEnvelope} label="Email address" value={profile.email}/><div className="divider my-0"/><ProfileDetail icon={faCalendarDays} label="Member since" value={formatDate(profile.createdAt)}/><div className="divider my-0"/><ProfileDetail icon={faClock} label="Session expires" value={formatDate(profile.sessionExpiresAt)}/></div></section>

        <section className="card border border-base-300 bg-base-100 shadow-sm"><div className="card-body p-3"><button className="flex min-h-14 items-center gap-4 rounded-box px-3 text-left hover:bg-base-200" onClick={() => navigate('/profile/sessions')}><span className="grid size-10 place-items-center rounded-full bg-primary/10 text-primary"><FontAwesomeIcon icon={faLaptop}/></span><span className="flex-1"><strong className="block text-sm">Manage sessions</strong><small className="text-base-content/45">Review devices signed into your account</small></span><FontAwesomeIcon icon={faArrowRight} className="text-base-content/30"/></button></div></section>

        <section className="card border border-base-300 bg-base-100 shadow-sm"><div className="card-body gap-4 p-5"><div><h2 className="text-sm font-semibold">Appearance</h2><p className="mt-1 text-xs leading-5 text-base-content/50">Choose how 0xSocial looks on this device.</p></div><ThemeSelector/></div></section>

        <button className="btn btn-outline btn-error" disabled={loggingOut} onClick={() => void logout()}>{loggingOut ? <span className="loading loading-spinner loading-sm"/> : <FontAwesomeIcon icon={faRightFromBracket}/>}Log out</button>
      </div>
    </AuthenticatedLayout>
  )
}
