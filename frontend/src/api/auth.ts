import { request } from './client'
export { ApiError } from './client'

export type Profile = {
  id: number
  email: string
  status: string
  handle: string
  displayName: string
  bio: string
  avatarUrl: string | null
  postCount: number
  followerCount: number
  followingCount: number
  createdAt: string
  updatedAt: string
  sessionExpiresAt: string
}

export type UpdateProfileInput = {
  displayName: string
  bio: string
}

export type RegisterInput = {
  email: string
  password: string
  handle: string
}

export type UserSession = {
  id: number
  current: boolean
  userAgent: string
  ipAddress: string
  authMethod: string
  riskLevel: string
  createdAt: string
  lastSeenAt: string
  expiresAt: string
}

export const authApi = {
  profile: (signal?: AbortSignal) => request<Profile>('/api/auth/me/profile', { method: 'GET', signal }),
  updateProfile: (input: UpdateProfileInput) => request<Profile>('/api/auth/me/profile', { method: 'PATCH', body: input, authenticated: true }),
  register: (input: RegisterInput) => request<void>('/api/auth/register', { body: input }),
  login: (email: string, password: string) => request<void>('/api/auth/login', { body: { email, password } }),
  logout: () => request<void>('/api/auth/logout', { authenticated: true }),
  verifyEmail: (token: string) => request<void>('/api/auth/verify-email', { body: { token } }),
  resendVerification: (email: string) => request<void>('/api/auth/resend-verification', { body: { email } }),
  forgotPassword: (email: string) => request<void>('/api/auth/forgot-password', { body: { email } }),
  resetPassword: (token: string, password: string) => request<void>('/api/auth/reset-password', { body: { token, password } }),
  sessions: () => request<{ sessions: UserSession[] }>('/api/auth/me/sessions', { method: 'GET', authenticated: true }),
  revokeSession: (sessionId: number) => request<void>(`/api/auth/me/sessions/${sessionId}`, { method: 'DELETE', authenticated: true }),
}
