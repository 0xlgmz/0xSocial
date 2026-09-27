import { useCallback, useEffect, useRef, useState } from 'react'
import type { ReactNode } from 'react'
import { ApiError, authApi } from '../../api/auth'
import type { Profile, UpdateProfileInput } from '../../api/auth'
import { sessionExpiredEvent } from '../../api/client'
import { navigate } from '../../lib/routes'
import { AuthContext } from './auth-context'
import type { AuthStatus } from './auth-context'

export function AuthProvider({ children }: { children: ReactNode }) {
  const [status, setStatus] = useState<AuthStatus>('checking')
  const [profile, setProfile] = useState<Profile | null>(null)
  const initialCheckStarted = useRef(false)

  const checkSession = useCallback(async () => {
    try {
      const user = await authApi.profile()
      setProfile(user)
      setStatus('authenticated')
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        setProfile(null)
        setStatus('public')
      } else {
        setStatus('error')
      }
    }
  }, [])

  const refreshProfile = useCallback(async () => {
    const user = await authApi.profile()
    setProfile(user)
    setStatus('authenticated')
  }, [])

  const updateProfile = useCallback(async (input: UpdateProfileInput) => {
    const updatedProfile = await authApi.updateProfile(input)
    setProfile(updatedProfile)
  }, [])

  const clearSession = useCallback(() => {
    setProfile(null)
    setStatus('public')
  }, [])

  const logout = useCallback(async () => {
    await authApi.logout()
    clearSession()
    navigate('/login')
  }, [clearSession])

  useEffect(() => {
    if (initialCheckStarted.current) return
    initialCheckStarted.current = true
    void checkSession()
  }, [checkSession])

  useEffect(() => {
    const handleSessionExpired = () => {
      clearSession()
      const returnTo = `${window.location.pathname}${window.location.search}`
      navigate('/login', { returnTo })
    }
    window.addEventListener(sessionExpiredEvent, handleSessionExpired)
    return () => window.removeEventListener(sessionExpiredEvent, handleSessionExpired)
  }, [clearSession])

  return (
    <AuthContext.Provider value={{ status, profile, checkSession, refreshProfile, updateProfile, logout, clearSession }}>
      {children}
    </AuthContext.Provider>
  )
}
