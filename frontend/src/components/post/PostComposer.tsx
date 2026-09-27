import { faImage, faPaperPlane, faXmark } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { useEffect, useMemo, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import { ApiError } from '../../api/client'
import { socialApi } from '../../api/social'
import type { SocialPost } from '../../api/social'
import { useCountdown } from '../../hooks/useCountdown'
import { MAX_IMAGE_INPUT_SIZE } from '../../lib/imageUpload'
import { Alert } from '../ui/Alert'
import { ProfileAvatar } from '../user/ProfileAvatar'

type PostComposerProps = {
  displayName: string
  handle: string
  avatarUrl: string | null
  onCreated: (post: SocialPost) => void
}

export function PostComposer({ displayName, handle, avatarUrl, onCreated }: PostComposerProps) {
  const [content, setContent] = useState('')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const [images, setImages] = useState<File[]>([])
  const [cooldown, setCooldown] = useCountdown()
  const fileInput = useRef<HTMLInputElement>(null)
  const previews = useMemo(() => images.map(file => ({ file, url: URL.createObjectURL(file) })), [images])
  const length = Array.from(content.trim()).length
  const valid = length > 0 && length <= 2000

  useEffect(() => () => previews.forEach(preview => URL.revokeObjectURL(preview.url)), [previews])

  function selectImages(files: FileList | null) {
    const selected = Array.from(files ?? [])
    setError('')
    if (selected.length > 4) { setError('You can attach up to four images.'); if (fileInput.current) fileInput.current.value = ''; return }
    if (selected.some(file => file.size > MAX_IMAGE_INPUT_SIZE)) { setError('Each original image must be 25 MB or smaller.'); if (fileInput.current) fileInput.current.value = ''; return }
    const allowedTypes = new Set(['image/jpeg', 'image/png', 'image/webp'])
    if (selected.some(file => !allowedTypes.has(file.type))) { setError('Images must be JPEG, PNG, or WebP files.'); if (fileInput.current) fileInput.current.value = ''; return }
    setImages(selected)
  }

  function removeImage(index: number) {
    setImages(current => current.filter((_, imageIndex) => imageIndex !== index))
    if (fileInput.current) fileInput.current.value = ''
  }

  async function submit(event: FormEvent) {
    event.preventDefault()
    if (!valid || busy || cooldown > 0) return
    setBusy(true); setError('')
    try {
      const mediaIds = await Promise.all(images.map(file => socialApi.uploadImage(file)))
      const post = await socialApi.createPost(content, mediaIds)
      setContent('')
      setImages([])
      if (fileInput.current) fileInput.current.value = ''
      onCreated(post)
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : 'Unable to create your post.')
      if (reason instanceof ApiError && reason.status === 429) setCooldown(reason.retryAfter ?? 30)
    } finally {
      setBusy(false)
    }
  }

  return (
    <section className="card border border-base-300 bg-base-100 shadow-sm" aria-labelledby="create-post-title">
      <form className="card-body gap-3 p-4 sm:p-5" onSubmit={submit}>
        <div className="flex gap-3">
          <ProfileAvatar displayName={displayName} fallbackText={`@${handle}`} avatarUrl={avatarUrl} size="sm"/>
          <div className="min-w-0 flex-1">
            <h1 id="create-post-title" className="sr-only">Create a post</h1>
            <textarea className={`textarea min-h-24 w-full resize-none border-0 bg-transparent p-1 text-base focus:outline-none ${length > 2000 ? 'text-error' : ''}`} value={content} onChange={event => setContent(event.target.value)} placeholder="What’s happening?" aria-invalid={length > 2000} disabled={busy}/>
          </div>
        </div>
        {previews.length > 0 && <div className={`grid gap-2 overflow-hidden rounded-xl ${previews.length > 1 ? 'grid-cols-2' : 'grid-cols-1'}`}>{previews.map((preview, index) => <div key={`${preview.file.name}-${preview.file.lastModified}-${index}`} className="group relative overflow-hidden rounded-xl bg-base-200"><img src={preview.url} alt={`Selected image ${index + 1}`} className="h-44 w-full object-cover sm:h-56"/><button className="btn btn-circle btn-sm absolute right-2 top-2 border border-base-300 bg-base-100/90 shadow-sm" type="button" onClick={() => removeImage(index)} disabled={busy} aria-label={`Remove selected image ${index + 1}`}><FontAwesomeIcon icon={faXmark}/></button></div>)}</div>}
        {error && <Alert type="error">{error}</Alert>}
        <div className="flex items-center justify-between border-t border-base-300 pt-3">
          <div className="flex items-center gap-3"><label className={`btn btn-ghost btn-circle btn-sm text-primary ${busy ? 'btn-disabled' : ''}`} aria-label="Attach images"><FontAwesomeIcon icon={faImage}/><input ref={fileInput} className="sr-only" type="file" accept="image/jpeg,image/png,image/webp" multiple disabled={busy} onChange={event => selectImages(event.target.files)}/></label><span className="text-xs text-base-content/40">{images.length}/4 images</span></div>
          <div className="flex items-center gap-3"><span className={`text-xs tabular-nums ${length > 2000 ? 'font-semibold text-error' : 'text-base-content/40'}`}>{length} / 2000</span><button className="btn btn-primary btn-sm" type="submit" disabled={!valid || busy || cooldown > 0}>{busy ? <><span className="loading loading-spinner loading-xs"/>Uploading…</> : cooldown > 0 ? `Try again in ${cooldown}s` : <><FontAwesomeIcon icon={faPaperPlane}/>Post</>}</button></div>
        </div>
      </form>
    </section>
  )
}
