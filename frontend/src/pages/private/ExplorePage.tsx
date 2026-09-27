import { faArrowRotateRight, faCompass, faImages, faXmark } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { useCallback, useEffect, useState } from 'react'
import { ApiError } from '../../api/client'
import { socialApi } from '../../api/social'
import type { SocialPost } from '../../api/social'
import { AuthenticatedLayout } from '../../components/layout/AuthenticatedLayout'
import { PostCard } from '../../components/post/PostCard'
import { Alert } from '../../components/ui/Alert'
import { useAuth } from '../../features/auth/useAuth'
import { navigate } from '../../lib/routes'

const pageSize = 20

export function ExplorePage() {
  const { profile, refreshProfile, clearSession } = useAuth()
  const [posts, setPosts] = useState<SocialPost[]>([])
  const [nextCursor, setNextCursor] = useState<number>()
  const [loading, setLoading] = useState(true)
  const [loadingMore, setLoadingMore] = useState(false)
  const [error, setError] = useState('')
  const [hasMore, setHasMore] = useState(false)
  const [selectedPost, setSelectedPost] = useState<SocialPost | null>(null)

  const handleError = useCallback((reason: unknown) => {
    if (reason instanceof ApiError && reason.status === 401) {
      clearSession()
      navigate('/login', { returnTo: '/explore' })
      return
    }
    setError(reason instanceof Error ? reason.message : 'Unable to load Explore.')
  }, [clearSession])

  useEffect(() => {
    const controller = new AbortController()
    socialApi.getExplore(undefined, pageSize, controller.signal)
      .then(page => {
        setPosts(page.posts)
        setNextCursor(page.nextCursor)
        setHasMore(page.posts.length === pageSize && page.nextCursor !== undefined)
      })
      .catch(reason => { if (!(reason instanceof DOMException && reason.name === 'AbortError')) handleError(reason) })
      .finally(() => setLoading(false))
    return () => controller.abort()
  }, [handleError])

  useEffect(() => {
    if (!selectedPost) return
    const closeOnEscape = (event: KeyboardEvent) => { if (event.key === 'Escape') setSelectedPost(null) }
    window.addEventListener('keydown', closeOnEscape)
    return () => window.removeEventListener('keydown', closeOnEscape)
  }, [selectedPost])

  async function loadMore() {
    if (!nextCursor || loadingMore) return
    setLoadingMore(true); setError('')
    try {
      const page = await socialApi.getExplore(nextCursor, pageSize)
      setPosts(current => [...current, ...page.posts])
      setNextCursor(page.nextCursor)
      setHasMore(page.posts.length === pageSize && page.nextCursor !== undefined)
    } catch (reason) { handleError(reason) }
    finally { setLoadingMore(false) }
  }

  async function remove(postId: number) {
    await socialApi.deletePost(postId)
    setPosts(current => current.filter(post => post.id !== postId))
    setSelectedPost(current => current?.id === postId ? null : current)
    void refreshProfile()
  }

  if (!profile) return null

  return (
    <AuthenticatedLayout title="Explore" route="/explore">
      <div className="flex flex-col gap-4">
        <div className="px-1"><h1 className="text-2xl font-semibold tracking-[-.03em]">Explore</h1><p className="mt-1 text-sm text-base-content/50">Discover the latest posts from across 0xSocial.</p></div>
        {error && <Alert type="error">{error}</Alert>}
        {loading ? <div className="grid place-items-center py-20"><span className="loading loading-spinner loading-lg text-primary"/></div> : posts.length === 0 ? <section className="card border border-base-300 bg-base-100 shadow-sm"><div className="card-body items-center py-16 text-center"><span className="grid size-16 place-items-center rounded-full bg-primary/10 text-2xl text-primary"><FontAwesomeIcon icon={faCompass}/></span><h2 className="mt-3 text-xl font-semibold">Nothing to explore yet</h2><p className="max-w-sm text-sm leading-6 text-base-content/50">Be the first person to share something with the community.</p><button className="btn btn-primary btn-sm mt-2" onClick={() => navigate('/create')}>Create post</button></div></section> : <section className="overflow-hidden rounded-box border border-base-300 bg-base-300 shadow-sm" aria-label="Global posts"><div className="grid grid-cols-3 gap-0.5">{posts.map(post => <button key={post.id} className="group relative aspect-square overflow-hidden bg-base-100 text-left focus-visible:z-10 focus-visible:outline-2 focus-visible:outline-primary" onClick={() => setSelectedPost(post)} aria-label={`Open post by @${post.handle}`}>{post.images.length > 0 ? <><img src={post.images[0]} alt="" loading="lazy" className="size-full object-cover transition-transform duration-300 group-hover:scale-[1.04]"/>{post.images.length > 1 && <span className="absolute right-2 top-2 grid size-7 place-items-center rounded-full bg-neutral/75 text-xs text-neutral-content"><FontAwesomeIcon icon={faImages}/></span>}<span className="absolute inset-x-0 bottom-0 truncate bg-gradient-to-t from-black/70 to-transparent px-3 pb-2 pt-8 text-xs font-medium text-white opacity-0 transition-opacity group-hover:opacity-100">@{post.handle}</span></> : <div className="flex size-full flex-col justify-between p-3 sm:p-4"><p className="line-clamp-[7] whitespace-pre-wrap break-words text-xs leading-5 text-base-content/75 sm:text-sm">{post.content}</p><span className="truncate text-[.65rem] text-base-content/40">@{post.handle}</span></div>}</button>)}</div></section>}
        {hasMore && <button className="btn btn-outline mx-auto" onClick={() => void loadMore()} disabled={loadingMore}>{loadingMore ? <span className="loading loading-spinner loading-sm"/> : <FontAwesomeIcon icon={faArrowRotateRight}/>}Load more</button>}
      </div>
      {selectedPost && <div className="modal modal-open" role="dialog" aria-modal="true" aria-label={`Post by @${selectedPost.handle}`}><div className="modal-box max-h-[88vh] max-w-3xl overflow-y-auto p-3 sm:p-5"><button className="btn btn-circle btn-sm absolute right-4 top-4 z-20 border border-base-300 bg-base-100 shadow-sm" onClick={() => setSelectedPost(null)} aria-label="Close post"><FontAwesomeIcon icon={faXmark}/></button><PostCard post={selectedPost} currentHandle={profile.handle} onDelete={remove}/></div><button className="modal-backdrop" onClick={() => setSelectedPost(null)} aria-label="Close post"/></div>}
    </AuthenticatedLayout>
  )
}
