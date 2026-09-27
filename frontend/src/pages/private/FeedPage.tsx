import { faArrowRotateRight, faCompass } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { useCallback, useEffect, useState } from 'react'
import { ApiError } from '../../api/client'
import { socialApi } from '../../api/social'
import type { SocialPost } from '../../api/social'
import { AuthenticatedLayout } from '../../components/layout/AuthenticatedLayout'
import { PostCard } from '../../components/post/PostCard'
import { FeedAdCard } from '../../components/ads/FeedAdCard'
import { Alert } from '../../components/ui/Alert'
import { useAuth } from '../../features/auth/useAuth'
import { interleaveFeedAds } from '../../features/monetization/interleaveFeedAds'
import { useMonetization } from '../../features/monetization/useMonetization'
import { navigate } from '../../lib/routes'

const pageSize = 20

export function FeedPage() {
  const { profile, refreshProfile, clearSession } = useAuth()
  const { config, canRequestAds } = useMonetization()
  const [posts, setPosts] = useState<SocialPost[]>([])
  const [nextCursor, setNextCursor] = useState<number>()
  const [loading, setLoading] = useState(true)
  const [loadingMore, setLoadingMore] = useState(false)
  const [error, setError] = useState('')
  const [hasMore, setHasMore] = useState(false)

  const handleError = useCallback((reason: unknown) => {
    if (reason instanceof ApiError && reason.status === 401) {
      clearSession()
      navigate('/login', { returnTo: '/feed' })
      return
    }
    setError(reason instanceof Error ? reason.message : 'Unable to load your feed.')
  }, [clearSession])

  useEffect(() => {
    const controller = new AbortController()
    socialApi.getFeed(undefined, pageSize, controller.signal)
      .then(page => {
        setPosts(page.posts)
        setNextCursor(page.nextCursor)
        setHasMore(page.posts.length === pageSize && page.nextCursor !== undefined)
      })
      .catch(reason => { if (!(reason instanceof DOMException && reason.name === 'AbortError')) handleError(reason) })
      .finally(() => setLoading(false))
    return () => controller.abort()
  }, [handleError])

  async function loadMore() {
    if (!nextCursor || loadingMore) return
    setLoadingMore(true); setError('')
    try {
      const page = await socialApi.getFeed(nextCursor, pageSize)
      setPosts(current => [...current, ...page.posts])
      setNextCursor(page.nextCursor)
      setHasMore(page.posts.length === pageSize && page.nextCursor !== undefined)
    } catch (reason) { handleError(reason) }
    finally { setLoadingMore(false) }
  }

  async function remove(postId: number) {
    await socialApi.deletePost(postId)
    setPosts(current => current.filter(post => post.id !== postId))
    void refreshProfile()
  }

  if (!profile) return null

  return (
    <AuthenticatedLayout title="Feed" route="/feed">
      <div className="flex flex-col gap-4">
        {error && <Alert type="error">{error}</Alert>}
        {loading ? (
          <div className="grid place-items-center py-20"><span className="loading loading-spinner loading-lg text-primary"/></div>
        ) : posts.length === 0 ? (
          <section className="card border border-base-300 bg-base-100 shadow-sm"><div className="card-body items-center py-16 text-center"><span className="grid size-16 place-items-center rounded-full bg-primary/10 text-2xl text-primary"><FontAwesomeIcon icon={faCompass}/></span><h2 className="mt-3 text-xl font-semibold">Your feed is empty</h2><p className="max-w-sm text-sm leading-6 text-base-content/50">Follow some people or create your first post.</p><div className="mt-2 flex gap-2"><button className="btn btn-outline btn-sm" onClick={() => navigate('/explore')}>Explore people</button><button className="btn btn-primary btn-sm" onClick={() => navigate('/create')}>Create post</button></div></div></section>
        ) : (
          <div className="flex flex-col gap-3">{canRequestAds
            ? interleaveFeedAds(
                posts,
                config.placements.feed,
                post => <PostCard key={`post-${post.id}`} post={post} currentHandle={profile.handle} onDelete={remove}/>,
                adIndex => <FeedAdCard key={`feed-ad-${adIndex}`} clientId={config.clientId} placement={config.placements.feed}/>,
              )
            : posts.map(post => <PostCard key={`post-${post.id}`} post={post} currentHandle={profile.handle} onDelete={remove}/>)}
          </div>
        )}
        {hasMore && <button className="btn btn-outline mx-auto" onClick={() => void loadMore()} disabled={loadingMore}>{loadingMore ? <span className="loading loading-spinner loading-sm"/> : <FontAwesomeIcon icon={faArrowRotateRight}/>}Load more</button>}
      </div>
    </AuthenticatedLayout>
  )
}
