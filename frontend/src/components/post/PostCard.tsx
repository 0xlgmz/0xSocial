import { faEllipsis, faFlag, faTrashCan, faXmark } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { useState } from 'react'
import { socialApi } from '../../api/social'
import type { ReportReason, SocialPost } from '../../api/social'
import { formatDate } from '../../lib/format'
import { navigateToProfile } from '../../lib/routes'
import { Alert } from '../ui/Alert'
import { ProfileAvatar } from '../user/ProfileAvatar'

type PostCardProps = { post: SocialPost; currentHandle?: string; onDelete?: (postId: number) => Promise<void> }

export function PostCard({ post, currentHandle, onDelete }: PostCardProps) {
  const [deleting, setDeleting] = useState(false)
  const [error, setError] = useState('')
  const [reportOpen, setReportOpen] = useState(false)
  const [reporting, setReporting] = useState(false)
  const [reported, setReported] = useState(false)
  const [reason, setReason] = useState<ReportReason>('spam')
  const ownPost = currentHandle === post.handle && onDelete

  async function remove() {
    if (!onDelete || deleting) return
    setDeleting(true); setError('')
    try { await onDelete(post.id) }
    catch (reason) { setError(reason instanceof Error ? reason.message : 'Unable to delete this post.'); setDeleting(false) }
  }

  async function report() {
    if (reporting || reported) return
    setReporting(true); setError('')
    try {
      await socialApi.reportPost(post.id, reason)
      setReported(true)
    } catch (reportError) {
      setError(reportError instanceof Error ? reportError.message : 'Unable to submit this report.')
    } finally { setReporting(false) }
  }

  return (
    <article className="card border border-base-300 bg-base-100 shadow-sm">
      <div className="card-body gap-3 p-4 sm:p-5">
        <div className="flex items-start gap-3">
          <button className="rounded-full" onClick={() => navigateToProfile(post.handle)} aria-label={`View @${post.handle}'s profile`}><ProfileAvatar displayName={post.displayName} fallbackText={`@${post.handle}`} avatarUrl={post.avatarUrl} size="sm"/></button>
          <div className="min-w-0 flex-1">
            <button className="block max-w-full text-left" onClick={() => navigateToProfile(post.handle)}><strong className="block truncate text-sm">{post.displayName || `@${post.handle}`}</strong><span className="block truncate text-xs text-base-content/45">@{post.handle}</span></button>
          </div>
          <details className="dropdown dropdown-end">
            <summary className="btn btn-ghost btn-circle btn-sm list-none text-base-content/40" aria-label="Post options"><FontAwesomeIcon icon={faEllipsis}/></summary>
            <ul className="menu dropdown-content z-20 mt-1 w-44 rounded-box border border-base-300 bg-base-100 p-2 shadow-lg">
              {!ownPost && <li><button onClick={() => setReportOpen(true)} disabled={reported}><FontAwesomeIcon icon={faFlag}/>{reported ? 'Report received' : 'Report post'}</button></li>}
              {ownPost && <li><button className="text-error" onClick={() => void remove()} disabled={deleting}>{deleting ? <span className="loading loading-spinner loading-xs"/> : <FontAwesomeIcon icon={faTrashCan}/>}Delete post</button></li>}
            </ul>
          </details>
        </div>
        <p className="whitespace-pre-wrap break-words text-sm leading-6 text-base-content/85">{post.content}</p>
        {post.images.length > 0 && <div className={`grid gap-1 overflow-hidden rounded-xl ${post.images.length > 1 ? 'grid-cols-2' : 'grid-cols-1'}`}>{post.images.map((image, index) => <img key={`${image}-${index}`} src={image} alt="" loading="lazy" className="max-h-96 h-full w-full object-cover"/>)}</div>}
        <time className="text-[.7rem] text-base-content/40" dateTime={post.createdAt}>{formatDate(post.createdAt)}</time>
        {error && <Alert type="error">{error}</Alert>}
      </div>
      {reportOpen && <div className="modal modal-open" role="dialog" aria-modal="true" aria-labelledby={`report-${post.id}-title`}><div className="modal-box"><button className="btn btn-ghost btn-circle btn-sm absolute right-3 top-3" onClick={() => setReportOpen(false)} aria-label="Close report dialog"><FontAwesomeIcon icon={faXmark}/></button>{reported ? <><h2 id={`report-${post.id}-title`} className="text-xl font-semibold">Report received</h2><p className="mt-3 text-sm leading-6 text-base-content/60">Thanks for letting us know. We’ll review the report without exposing internal moderation decisions.</p><div className="modal-action"><button className="btn btn-primary" onClick={() => setReportOpen(false)}>Done</button></div></> : <><h2 id={`report-${post.id}-title`} className="text-xl font-semibold">Report post</h2><p className="mt-2 text-sm text-base-content/55">What best describes the issue?</p><select className="select mt-5 w-full" value={reason} onChange={event => setReason(event.target.value as ReportReason)}><option value="spam">Spam or scam</option><option value="harassment">Harassment</option><option value="hate_speech">Hateful conduct</option><option value="violence">Violence or threats</option><option value="sexual_content">Sexual content</option><option value="self_harm">Self-harm</option><option value="false_information">Harmful false information</option><option value="other">Something else</option></select><div className="modal-action"><button className="btn btn-ghost" onClick={() => setReportOpen(false)}>Cancel</button><button className="btn btn-primary" onClick={() => void report()} disabled={reporting}>{reporting && <span className="loading loading-spinner loading-sm"/>}Submit report</button></div></>}</div><button className="modal-backdrop" onClick={() => setReportOpen(false)} aria-label="Close report dialog"/></div>}
    </article>
  )
}
