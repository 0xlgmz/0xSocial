import { faArrowLeft, faPen, faRotateRight, faTriangleExclamation, faUserPlus, faUserSlash } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { useEffect, useState } from 'react'
import { ApiError } from '../../api/client'
import { profileApi } from '../../api/profile'
import type { PublicProfile } from '../../api/profile'
import { socialApi } from '../../api/social'
import type { FollowListKind, Relationship, SocialPost } from '../../api/social'
import { Logo } from '../../components/brand/Logo'
import { AuthenticatedLayout } from '../../components/layout/AuthenticatedLayout'
import { Alert } from '../../components/ui/Alert'
import { FollowListDialog } from '../../components/user/FollowListDialog'
import { ProfileView } from '../../components/user/ProfileView'
import { useAuth } from '../../features/auth/useAuth'
import { useCountdown } from '../../hooks/useCountdown'
import { navigate } from '../../lib/routes'

const pageSize = 20
type ProfileState = { kind: 'loading' } | { kind: 'loaded'; profile: PublicProfile } | { kind: 'not-found' } | { kind: 'error' }

function PublicProfileHeader({ showActions }: { showActions: boolean }) {
  return <header className="border-b border-base-300 bg-base-100"><div className="mx-auto flex h-16 max-w-3xl items-center justify-between px-5"><Logo/>{showActions && <div className="flex gap-2"><button className="btn btn-ghost btn-sm" onClick={() => navigate('/login')}>Log in</button><button className="btn btn-primary btn-sm" onClick={() => navigate('/register')}>Join 0xSocial</button></div>}</div></header>
}

function ProfileSkeleton() {
  return <div className="overflow-hidden rounded-box border border-base-300 bg-base-100 shadow-sm"><div className="px-6 py-10"><div className="skeleton mx-auto h-8 w-36"/><div className="mt-8 flex items-center gap-8"><div className="skeleton size-24 rounded-full"/><div className="grid flex-1 grid-cols-3 gap-4"><div className="skeleton h-12"/><div className="skeleton h-12"/><div className="skeleton h-12"/></div></div><div className="skeleton mt-6 h-4 w-40"/><div className="skeleton mt-2 h-4 w-3/5"/></div><div className="grid grid-cols-3 gap-0.5 border-t border-base-300"><div className="skeleton aspect-square rounded-none"/><div className="skeleton aspect-square rounded-none"/><div className="skeleton aspect-square rounded-none"/></div></div>
}

export function PublicProfilePage({ handle }: { handle: string }) {
  const { status, clearSession, refreshProfile } = useAuth()
  const [state, setState] = useState<ProfileState>({ kind: 'loading' })
  const [posts, setPosts] = useState<SocialPost[]>([])
  const [postsLoading, setPostsLoading] = useState(true)
  const [postsError, setPostsError] = useState('')
  const [nextCursor, setNextCursor] = useState<number>()
  const [hasMore, setHasMore] = useState(false)
  const [loadingMore, setLoadingMore] = useState(false)
  const [relationship, setRelationship] = useState<Relationship | null>(null)
  const [actionBusy, setActionBusy] = useState(false)
  const [actionError, setActionError] = useState('')
  const [followCooldown, setFollowCooldown] = useCountdown()
  const [list, setList] = useState<FollowListKind | null>(null)
  const [attempt, setAttempt] = useState(0)

  useEffect(() => {
    const controller = new AbortController()
    Promise.all([profileApi.getByHandle(handle, controller.signal), socialApi.getProfilePosts(handle, undefined, pageSize, controller.signal)])
      .then(([profile, page]) => { setState({ kind: 'loaded', profile }); setPosts(page.posts); setNextCursor(page.nextCursor); setHasMore(page.posts.length === pageSize && page.nextCursor !== undefined) })
      .catch(error => {
        if (error instanceof DOMException && error.name === 'AbortError') return
        setState(error instanceof ApiError && error.status === 404 ? { kind: 'not-found' } : { kind: 'error' })
      })
      .finally(() => setPostsLoading(false))
    return () => controller.abort()
  }, [handle, attempt])

  useEffect(() => {
    if (status !== 'authenticated') return
    const controller = new AbortController()
    socialApi.getRelationship(handle, controller.signal)
      .then(setRelationship)
      .catch(reason => {
        if (reason instanceof DOMException && reason.name === 'AbortError') return
        if (reason instanceof ApiError && reason.status === 401) { clearSession(); navigate('/login', { returnTo: `/@${handle}` }); return }
        setActionError(reason instanceof Error ? reason.message : 'Unable to load follow status.')
      })
    return () => controller.abort()
  }, [status, handle, clearSession])

  function retry() { setState({ kind: 'loading' }); setPostsLoading(true); setPostsError(''); setAttempt(value => value + 1) }

  async function loadMore() {
    if (!nextCursor || loadingMore) return
    setLoadingMore(true); setPostsError('')
    try { const page = await socialApi.getProfilePosts(handle, nextCursor, pageSize); setPosts(current => [...current, ...page.posts]); setNextCursor(page.nextCursor); setHasMore(page.posts.length === pageSize && page.nextCursor !== undefined) }
    catch (reason) { setPostsError(reason instanceof Error ? reason.message : 'Unable to load more posts.') }
    finally { setLoadingMore(false) }
  }

  async function toggleFollow() {
    if (!relationship || relationship.isSelf || actionBusy || followCooldown > 0) return
    const wasFollowing = relationship.following
    const delta = wasFollowing ? -1 : 1
    setActionBusy(true); setActionError('')
    setRelationship(current => current ? { ...current, following: !wasFollowing } : current)
    setState(current => current.kind === 'loaded' ? { kind: 'loaded', profile: { ...current.profile, followerCount: Math.max(0, current.profile.followerCount + delta) } } : current)
    try { if (wasFollowing) await socialApi.unfollow(handle); else await socialApi.follow(handle); void refreshProfile() }
    catch (reason) {
      setRelationship(current => current ? { ...current, following: wasFollowing } : current)
      setState(current => current.kind === 'loaded' ? { kind: 'loaded', profile: { ...current.profile, followerCount: Math.max(0, current.profile.followerCount - delta) } } : current)
      if (reason instanceof ApiError && reason.status === 401) { clearSession(); navigate('/login', { returnTo: `/@${handle}` }); return }
      if (reason instanceof ApiError && reason.status === 429) setFollowCooldown(reason.retryAfter ?? 30)
      setActionError(reason instanceof Error ? reason.message : 'Unable to update follow status.')
    } finally { setActionBusy(false) }
  }

  let content
  if (state.kind === 'loading') content = <ProfileSkeleton/>
  else if (state.kind === 'loaded') {
    let action = null
    if (relationship?.isSelf) action = <button className="btn btn-outline btn-sm w-full sm:w-auto" onClick={() => navigate('/profile/edit')}><FontAwesomeIcon icon={faPen}/>Edit profile</button>
    else if (status === 'authenticated') action = <div className="flex flex-wrap items-center gap-2"><button className={`btn btn-sm ${relationship?.following ? 'btn-outline' : 'btn-primary'}`} onClick={() => void toggleFollow()} disabled={!relationship || actionBusy || followCooldown > 0}>{actionBusy || !relationship ? <span className="loading loading-spinner loading-xs"/> : followCooldown > 0 ? `Try again in ${followCooldown}s` : <>{!relationship.following && <FontAwesomeIcon icon={faUserPlus}/>} {relationship.following ? 'Following' : 'Follow'}</>}</button>{relationship?.followedBy && <span className="badge badge-soft">Follows you</span>}</div>
    content = <>{actionError && <div className="mb-4"><Alert type="error">{actionError}</Alert></div>}<ProfileView displayName={state.profile.displayName} handle={state.profile.handle} bio={state.profile.bio} avatarUrl={state.profile.avatarUrl} postCount={state.profile.postCount} followerCount={state.profile.followerCount} followingCount={state.profile.followingCount} posts={posts} postsLoading={postsLoading} postsError={postsError} hasMorePosts={hasMore} loadingMorePosts={loadingMore} action={action} onFollowers={() => setList('followers')} onFollowing={() => setList('following')} onLoadMorePosts={() => void loadMore()}/>{list && <FollowListDialog handle={handle} kind={list} onClose={() => setList(null)}/>}</>
  } else if (state.kind === 'not-found') content = <div className="card border border-base-300 bg-base-100 shadow-sm"><div className="card-body items-center py-14 text-center"><div className="grid size-20 place-items-center rounded-full bg-base-200 text-3xl text-base-content/35"><FontAwesomeIcon icon={faUserSlash}/></div><p className="mt-4 font-mono text-xs uppercase tracking-[.2em] text-primary">@{handle}</p><h1 className="text-3xl font-semibold tracking-[-.035em]">User not found</h1><p className="max-w-sm text-sm leading-6 text-base-content/50">This profile doesn’t exist, or the account is no longer available.</p><button className="btn btn-primary mt-4" onClick={() => navigate('/')}><FontAwesomeIcon icon={faArrowLeft}/>Back to 0xSocial</button></div></div>
  else content = <div className="card border border-base-300 bg-base-100 shadow-sm"><div className="card-body items-center py-14 text-center"><div className="grid size-20 place-items-center rounded-full bg-error/10 text-3xl text-error"><FontAwesomeIcon icon={faTriangleExclamation}/></div><h1 className="mt-4 text-3xl font-semibold tracking-[-.035em]">Couldn’t load this profile</h1><p className="max-w-sm text-sm leading-6 text-base-content/50">Something went wrong while connecting to 0xSocial. Please try again.</p><button className="btn btn-primary mt-4" onClick={retry}><FontAwesomeIcon icon={faRotateRight}/>Try again</button></div></div>

  if (status === 'authenticated') return <AuthenticatedLayout title={`@${handle}`} route={null}>{content}</AuthenticatedLayout>
  return <main className="min-h-screen bg-base-200"><PublicProfileHeader showActions={status === 'public'}/><section className="mx-auto max-w-2xl p-4 py-8 sm:p-8 sm:py-12">{content}</section></main>
}
