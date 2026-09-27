import { request } from './client'

export type PublicProfile = {
  handle: string
  displayName: string
  bio: string
  avatarUrl: string | null
  postCount: number
  followerCount: number
  followingCount: number
}

export const profileApi = {
  getByHandle: (handle: string, signal?: AbortSignal) =>
    request<PublicProfile>(`/api/profiles/${encodeURIComponent(handle)}`, {
      method: 'GET',
      signal,
    }),
}
