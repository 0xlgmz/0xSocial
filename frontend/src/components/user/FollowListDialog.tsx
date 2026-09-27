import { faArrowRotateRight, faXmark } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { useEffect, useState } from 'react'
import { socialApi } from '../../api/social'
import type { FollowListKind, SocialProfile } from '../../api/social'
import { navigateToProfile } from '../../lib/routes'
import { Alert } from '../ui/Alert'
import { ProfileAvatar } from './ProfileAvatar'

const pageSize = 20

type FollowListDialogProps = { handle: string; kind: FollowListKind; onClose: () => void }

export function FollowListDialog({ handle, kind, onClose }: FollowListDialogProps) {
  const [profiles, setProfiles] = useState<SocialProfile[]>([])
  const [nextCursor, setNextCursor] = useState<number>()
  const [loading, setLoading] = useState(true)
  const [loadingMore, setLoadingMore] = useState(false)
  const [hasMore, setHasMore] = useState(false)
  const [error, setError] = useState('')
  const title = kind === 'followers' ? 'Followers' : 'Following'

  useEffect(() => {
    const controller = new AbortController()
    socialApi.getProfiles(handle, kind, undefined, pageSize, controller.signal)
      .then(page => { setProfiles(page.profiles); setNextCursor(page.nextCursor); setHasMore(page.profiles.length === pageSize && page.nextCursor !== undefined) })
      .catch(reason => { if (!(reason instanceof DOMException && reason.name === 'AbortError')) setError(reason instanceof Error ? reason.message : `Unable to load ${kind}.`) })
      .finally(() => setLoading(false))
    return () => controller.abort()
  }, [handle, kind])

  async function loadMore() {
    if (!nextCursor || loadingMore) return
    setLoadingMore(true); setError('')
    try {
      const page = await socialApi.getProfiles(handle, kind, nextCursor, pageSize)
      setProfiles(current => [...current, ...page.profiles])
      setNextCursor(page.nextCursor)
      setHasMore(page.profiles.length === pageSize && page.nextCursor !== undefined)
    } catch (reason) { setError(reason instanceof Error ? reason.message : `Unable to load more ${kind}.`) }
    finally { setLoadingMore(false) }
  }

  function openProfile(profileHandle: string) { onClose(); navigateToProfile(profileHandle) }

  return (
    <div className="modal modal-open" role="dialog" aria-modal="true" aria-labelledby="follow-list-title">
      <div className="modal-box flex max-h-[80vh] max-w-lg flex-col border border-base-300 p-0">
        <div className="flex items-center justify-between border-b border-base-300 px-5 py-4"><div><h2 id="follow-list-title" className="font-semibold">{title}</h2><p className="text-xs text-base-content/45">@{handle}</p></div><button className="btn btn-ghost btn-circle btn-sm" onClick={onClose} aria-label={`Close ${title.toLowerCase()}`}><FontAwesomeIcon icon={faXmark}/></button></div>
        <div className="min-h-44 overflow-y-auto p-3">
          {error && <Alert type="error">{error}</Alert>}
          {loading ? <div className="grid place-items-center py-16"><span className="loading loading-spinner text-primary"/></div> : profiles.length === 0 ? <p className="py-16 text-center text-sm text-base-content/45">No {kind} yet.</p> : <div className="flex flex-col">{profiles.map(profile => <button key={profile.handle} className="flex items-center gap-3 rounded-box p-3 text-left hover:bg-base-200" onClick={() => openProfile(profile.handle)}><ProfileAvatar displayName={profile.displayName} fallbackText={`@${profile.handle}`} avatarUrl={profile.avatarUrl} size="sm"/><span className="min-w-0 flex-1"><strong className="block truncate text-sm">{profile.displayName || `@${profile.handle}`}</strong><span className="block truncate text-xs text-base-content/45">@{profile.handle}</span>{profile.bio && <span className="mt-1 block truncate text-xs text-base-content/55">{profile.bio}</span>}</span></button>)}</div>}
          {hasMore && <button className="btn btn-outline btn-sm mx-auto mt-3 flex" onClick={() => void loadMore()} disabled={loadingMore}>{loadingMore ? <span className="loading loading-spinner loading-xs"/> : <FontAwesomeIcon icon={faArrowRotateRight}/>}Load more</button>}
        </div>
      </div>
      <button className="modal-backdrop" onClick={onClose} aria-label={`Close ${title.toLowerCase()}`}/>
    </div>
  )
}
