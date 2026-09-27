import { faPen } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { useEffect, useState } from 'react'
import type { Profile } from '../../api/auth'
import { socialApi } from '../../api/social'
import type { FollowListKind, SocialPost } from '../../api/social'
import { AuthenticatedLayout } from '../../components/layout/AuthenticatedLayout'
import { FollowListDialog } from '../../components/user/FollowListDialog'
import { ProfileView } from '../../components/user/ProfileView'
import { useAuth } from '../../features/auth/useAuth'
import { navigate } from '../../lib/routes'

const pageSize = 20

export function ProfilePage({ profile }: { profile: Profile }) {
  const { refreshProfile } = useAuth()
  const [posts, setPosts] = useState<SocialPost[]>([])
  const [nextCursor, setNextCursor] = useState<number>()
  const [loading, setLoading] = useState(true)
  const [loadingMore, setLoadingMore] = useState(false)
  const [hasMore, setHasMore] = useState(false)
  const [error, setError] = useState('')
  const [list, setList] = useState<FollowListKind | null>(null)

  useEffect(() => {
    const controller = new AbortController()
    socialApi.getProfilePosts(profile.handle, undefined, pageSize, controller.signal)
      .then(page => { setPosts(page.posts); setNextCursor(page.nextCursor); setHasMore(page.posts.length === pageSize && page.nextCursor !== undefined) })
      .catch(reason => { if (!(reason instanceof DOMException && reason.name === 'AbortError')) setError(reason instanceof Error ? reason.message : 'Unable to load posts.') })
      .finally(() => setLoading(false))
    return () => controller.abort()
  }, [profile.handle])

  async function loadMore() {
    if (!nextCursor || loadingMore) return
    setLoadingMore(true); setError('')
    try { const page = await socialApi.getProfilePosts(profile.handle, nextCursor, pageSize); setPosts(current => [...current, ...page.posts]); setNextCursor(page.nextCursor); setHasMore(page.posts.length === pageSize && page.nextCursor !== undefined) }
    catch (reason) { setError(reason instanceof Error ? reason.message : 'Unable to load more posts.') }
    finally { setLoadingMore(false) }
  }

  async function remove(postId: number) {
    await socialApi.deletePost(postId)
    setPosts(current => current.filter(post => post.id !== postId))
    await refreshProfile()
  }

  return (
    <AuthenticatedLayout title="Profile" route="/profile">
      <ProfileView displayName={profile.displayName} handle={profile.handle} bio={profile.bio} avatarUrl={profile.avatarUrl} postCount={profile.postCount} followerCount={profile.followerCount} followingCount={profile.followingCount} posts={posts} postsLoading={loading} postsError={error} hasMorePosts={hasMore} loadingMorePosts={loadingMore} currentHandle={profile.handle} action={<button className="btn btn-outline btn-sm w-full sm:w-auto" onClick={() => navigate('/profile/edit')}><FontAwesomeIcon icon={faPen}/>Edit profile</button>} onFollowers={() => setList('followers')} onFollowing={() => setList('following')} onLoadMorePosts={() => void loadMore()} onDeletePost={remove}/>
      {list && <FollowListDialog handle={profile.handle} kind={list} onClose={() => setList(null)}/>} 
    </AuthenticatedLayout>
  )
}
