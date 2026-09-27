import { request } from './client'
import { prepareImageForUpload } from '../lib/imageUpload'

export type SocialPost = {
  id: number
  handle: string
  displayName: string
  avatarUrl: string | null
  content: string
  createdAt: string
  updatedAt: string
  images: string[]
}

export type MediaUploadTicket = {
  mediaId: number
  uploadUrl: string
  method: 'PUT'
  headers: Record<string, string>
  expiresAt: string
}

export type PostsPage = { posts: SocialPost[]; nextCursor?: number }

export type SocialProfile = {
  handle: string
  displayName: string
  bio: string
  avatarUrl: string | null
  followedAt: string
}

export type ProfilesPage = { profiles: SocialProfile[]; nextCursor?: number }
export type Relationship = { isSelf: boolean; following: boolean; followedBy: boolean }
export type FollowListKind = 'followers' | 'following'
export type ReportReason = 'spam' | 'harassment' | 'hate_speech' | 'violence' | 'sexual_content' | 'self_harm' | 'false_information' | 'other'

function pageQuery(before?: number, limit = 20) {
  const query = new URLSearchParams({ limit: String(limit) })
  if (before !== undefined) query.set('before', String(before))
  return query.toString()
}

async function uploadImage(file: File): Promise<number> {
  const image = await prepareImageForUpload(file)
  const ticket = await request<MediaUploadTicket>('/api/media/uploads', {
    authenticated: true,
    body: { contentType: image.type, sizeBytes: image.size },
  })

  const response = await fetch(ticket.uploadUrl, {
    method: ticket.method,
    headers: ticket.headers,
    body: image,
    credentials: 'omit',
  })

  if (!response.ok) throw new Error('Unable to upload image.')
  return ticket.mediaId
}

export const socialApi = {
  uploadImage,
  createPost: (content: string, mediaIds: number[] = []) => request<SocialPost>('/api/posts', { body: { content, mediaIds }, authenticated: true }),
  deletePost: (postId: number) => request<void>(`/api/posts/${encodeURIComponent(postId)}`, { method: 'DELETE', authenticated: true }),
  reportPost: (postId: number, reason: ReportReason) => request<void>(`/api/posts/${encodeURIComponent(postId)}/reports`, { body: { reason }, authenticated: true }),
  getFeed: (before?: number, limit = 20, signal?: AbortSignal) => request<PostsPage>(`/api/feed?${pageQuery(before, limit)}`, { method: 'GET', signal, authenticated: true }),
  getExplore: (before?: number, limit = 20, signal?: AbortSignal) => request<PostsPage>(`/api/explore?${pageQuery(before, limit)}`, { method: 'GET', signal, authenticated: true }),
  getProfilePosts: (handle: string, before?: number, limit = 20, signal?: AbortSignal) => request<PostsPage>(`/api/profiles/${encodeURIComponent(handle)}/posts?${pageQuery(before, limit)}`, { method: 'GET', signal }),
  getRelationship: (handle: string, signal?: AbortSignal) => request<Relationship>(`/api/profiles/${encodeURIComponent(handle)}/relationship`, { method: 'GET', signal, authenticated: true }),
  follow: (handle: string) => request<void>(`/api/profiles/${encodeURIComponent(handle)}/follow`, { method: 'PUT', authenticated: true }),
  unfollow: (handle: string) => request<void>(`/api/profiles/${encodeURIComponent(handle)}/follow`, { method: 'DELETE', authenticated: true }),
  getProfiles: (handle: string, kind: FollowListKind, before?: number, limit = 20, signal?: AbortSignal) => request<ProfilesPage>(`/api/profiles/${encodeURIComponent(handle)}/${kind}?${pageQuery(before, limit)}`, { method: 'GET', signal }),
}
