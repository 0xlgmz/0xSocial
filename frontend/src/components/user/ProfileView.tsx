import { faArrowRotateRight, faCamera, faTrashCan } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { useState } from 'react'
import type { ReactNode } from 'react'
import type { SocialPost } from '../../api/social'
import { formatDate } from '../../lib/format'
import { Alert } from '../ui/Alert'
import { ProfileAvatar } from './ProfileAvatar'

type ProfileViewProps = {
  displayName: string
  handle: string
  bio: string
  avatarUrl: string | null
  postCount: number
  followerCount: number
  followingCount: number
  posts: SocialPost[]
  postsLoading?: boolean
  postsError?: string
  hasMorePosts?: boolean
  loadingMorePosts?: boolean
  currentHandle?: string
  action?: ReactNode
  onFollowers: () => void
  onFollowing: () => void
  onLoadMorePosts?: () => void
  onDeletePost?: (postId: number) => Promise<void>
}

function Stat({ value, label, onClick }: { value: number; label: string; onClick?: () => void }) {
  const content = <><strong className="block text-base font-semibold tabular-nums">{value.toLocaleString()}</strong><span className="mt-1 block text-xs text-base-content/55 sm:text-sm">{label}</span></>
  return onClick ? <button className="rounded-lg py-1 text-center hover:bg-base-200 focus-visible:outline-2 focus-visible:outline-primary" onClick={onClick}>{content}</button> : <div className="py-1 text-center">{content}</div>
}

export function ProfileView({ displayName, handle, bio, avatarUrl, postCount, followerCount, followingCount, posts, postsLoading = false, postsError = '', hasMorePosts = false, loadingMorePosts = false, currentHandle, action, onFollowers, onFollowing, onLoadMorePosts, onDeletePost }: ProfileViewProps) {
  const resolvedName = displayName.trim() || `@${handle}`
  const [deleteError, setDeleteError] = useState('')
  const [deletingId, setDeletingId] = useState<number>()

  async function remove(postId: number) {
    if (!onDeletePost || deletingId) return
    setDeletingId(postId); setDeleteError('')
    try { await onDeletePost(postId) }
    catch (reason) { setDeleteError(reason instanceof Error ? reason.message : 'Unable to delete this post.') }
    finally { setDeletingId(undefined) }
  }

  return (
    <section className="overflow-hidden rounded-box border border-base-300 bg-base-100 shadow-sm">
      <div className="px-5 pb-6 pt-7 sm:px-10 sm:pb-8 sm:pt-9">
        <h1 className="text-center text-2xl font-semibold tracking-[-.035em] sm:text-3xl">@{handle}</h1>
        <div className="mt-7 grid grid-cols-[6rem_1fr] items-center gap-5 sm:grid-cols-[7rem_1fr] sm:gap-10">
          <ProfileAvatar displayName={displayName} fallbackText={`@${handle}`} avatarUrl={avatarUrl}/>
          <div className="grid grid-cols-3 gap-1 sm:gap-3"><Stat value={postCount} label="posts"/><Stat value={followerCount} label="followers" onClick={onFollowers}/><Stat value={followingCount} label="following" onClick={onFollowing}/></div>
        </div>
        <div className="mt-5 max-w-lg"><h2 className="text-sm font-semibold">{resolvedName}</h2>{bio ? <p className="mt-1 whitespace-pre-wrap text-sm leading-6 text-base-content/70">{bio}</p> : <p className="mt-1 text-sm italic text-base-content/40">No bio added yet.</p>}</div>
        {action && <div className="mt-6">{action}</div>}
      </div>

      <div className="border-t border-base-300" aria-label={`${resolvedName}'s posts`}>
        {(postsError || deleteError) && <div className="p-4"><Alert type="error">{postsError || deleteError}</Alert></div>}
        {postsLoading ? <div className="grid min-h-56 place-items-center"><span className="loading loading-spinner text-primary"/></div> : posts.length > 0 ? (
          <div className="grid grid-cols-3 gap-0.5 bg-base-300">
            {posts.map(post => {
              const canDelete = currentHandle === post.handle && onDeletePost
              return <article key={post.id} className="group relative aspect-square overflow-hidden bg-base-100">{post.images.length > 0 ? <><img src={post.images[0]} alt="" loading="lazy" className="size-full object-cover transition-transform duration-300 group-hover:scale-[1.03]"/>{post.images.length > 1 && <span className="absolute bottom-2 right-2 rounded-full bg-neutral/80 px-2 py-1 text-[.65rem] font-semibold text-neutral-content">+{post.images.length - 1}</span>}</> : <div className="size-full p-3 sm:p-4"><p className="line-clamp-[7] whitespace-pre-wrap break-words text-xs leading-5 text-base-content/75 sm:text-sm">{post.content}</p><time className="absolute inset-x-3 bottom-3 truncate bg-base-100/90 text-[.6rem] text-base-content/40 sm:inset-x-4" dateTime={post.createdAt}>{formatDate(post.createdAt)}</time></div>}{canDelete && <button className="btn btn-circle btn-xs absolute right-2 top-2 border border-base-300 bg-base-100/90 text-base-content/45 opacity-100 shadow-sm sm:opacity-0 sm:group-hover:opacity-100" onClick={() => void remove(post.id)} disabled={deletingId === post.id} aria-label="Delete post">{deletingId === post.id ? <span className="loading loading-spinner loading-xs"/> : <FontAwesomeIcon icon={faTrashCan}/>}</button>}</article>
            })}
          </div>
        ) : <div className="grid min-h-56 place-items-center px-6 py-12 text-center"><div><span className="mx-auto grid size-14 place-items-center rounded-full border border-base-300 text-xl text-base-content/35"><FontAwesomeIcon icon={faCamera}/></span><h2 className="mt-4 text-sm font-semibold">No posts yet</h2><p className="mt-1 text-xs text-base-content/45">Posts will appear here in a three-column grid.</p></div></div>}
        {hasMorePosts && <div className="flex justify-center border-t border-base-300 p-4"><button className="btn btn-outline btn-sm" onClick={onLoadMorePosts} disabled={loadingMorePosts}>{loadingMorePosts ? <span className="loading loading-spinner loading-xs"/> : <FontAwesomeIcon icon={faArrowRotateRight}/>}Load more posts</button></div>}
      </div>
    </section>
  )
}
