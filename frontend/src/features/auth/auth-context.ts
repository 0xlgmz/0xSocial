import { createContext } from 'react'
import type { Profile, UpdateProfileInput } from '../../api/auth'

export type AuthStatus = 'checking' | 'authenticated' | 'public' | 'error'

export type AuthContextValue = {
  status: AuthStatus
  profile: Profile | null
  checkSession: () => Promise<void>
  refreshProfile: () => Promise<void>
  updateProfile: (input: UpdateProfileInput) => Promise<void>
  logout: () => Promise<void>
  clearSession: () => void
}

export const AuthContext = createContext<AuthContextValue | null>(null)
