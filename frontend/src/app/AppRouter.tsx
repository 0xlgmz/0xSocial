import { faBell } from '@fortawesome/free-solid-svg-icons'
import { useEffect, useState } from 'react'
import { ConnectionError } from '../components/layout/ConnectionError'
import { LoadingScreen } from '../components/layout/LoadingScreen'
import { useAuth } from '../features/auth/useAuth'
import { getDocumentTitle } from '../lib/documentTitle'
import { normalizeHandle } from '../lib/handle'
import { getCurrentRoute, legalRoutes, navigate, privateRoutes, publicRoutes, redirect } from '../lib/routes'
import type { LegalRoute, PrivateRoute, PublicRoute, ResolvedRoute } from '../lib/routes'
import { PageNotFound } from '../pages/PageNotFound'
import { FeedPage } from '../pages/private/FeedPage'
import { ExplorePage } from '../pages/private/ExplorePage'
import { CreatePostPage } from '../pages/private/CreatePostPage'
import { EditProfilePage } from '../pages/private/EditProfilePage'
import { PlaceholderPage } from '../pages/private/PlaceholderPage'
import { ProfilePage } from '../pages/private/ProfilePage'
import { SessionsPage } from '../pages/private/SessionsPage'
import { CheckEmailPage } from '../pages/public/CheckEmailPage'
import { ForgotPasswordPage } from '../pages/public/ForgotPasswordPage'
import { LoginPage } from '../pages/public/LoginPage'
import { PublicProfilePage } from '../pages/public/PublicProfilePage'
import { RegisterPage } from '../pages/public/RegisterPage'
import { ResetPasswordPage } from '../pages/public/ResetPasswordPage'
import { VerifyEmailPage } from '../pages/public/VerifyEmailPage'
import { WelcomePage } from '../pages/public/WelcomePage'
import { LegalPage } from '../pages/public/LegalPage'

export function AppRouter() {
  const [route, setRoute] = useState<ResolvedRoute>(getCurrentRoute)
  const { status, profile, checkSession, refreshProfile, updateProfile, logout, clearSession } = useAuth()

  useEffect(() => {
    const handleRoute = () => setRoute(getCurrentRoute())
    window.addEventListener('popstate', handleRoute)
    return () => window.removeEventListener('popstate', handleRoute)
  }, [])

  useEffect(() => {
    document.title = getDocumentTitle(route)
  }, [route])

  useEffect(() => {
    if (
      route.kind === 'profile'
      && status === 'authenticated'
      && profile
      && route.handle === normalizeHandle(profile.handle)
    ) {
      redirect('/profile')
      return
    }

    if (route.kind !== 'static') return
    if (status === 'authenticated' && publicRoutes.has(route.path as PublicRoute) && !legalRoutes.has(route.path as LegalRoute) && route.path !== '/verify-email') navigate('/feed')
    if (status === 'public' && privateRoutes.has(route.path as PrivateRoute)) navigate('/login')
  }, [status, profile, route])

  if (route.kind === 'profile') {
    const isOwnProfile = status === 'authenticated'
      && profile
      && route.handle === normalizeHandle(profile.handle)

    if (isOwnProfile) return <LoadingScreen/>
    return <PublicProfilePage key={route.handle} handle={route.handle}/>
  }
  if (route.kind === 'not-found') return <PageNotFound/>
  if (legalRoutes.has(route.path as LegalRoute)) return <LegalPage route={route.path as LegalRoute}/>

  if (status === 'checking') return <LoadingScreen/>
  if (status === 'error') return <ConnectionError onRetry={() => void checkSession()}/>

  if (status === 'public') {
    switch (route.path) {
      case '/register': return <RegisterPage/>
      case '/check-email': return <CheckEmailPage/>
      case '/verify-email': return <VerifyEmailPage onAuthenticated={refreshProfile}/>
      case '/login': return <LoginPage onAuthenticated={refreshProfile}/>
      case '/forgot-password': return <ForgotPasswordPage/>
      case '/reset-password': return <ResetPasswordPage/>
      default: return <WelcomePage/>
    }
  }

  if (!profile) return <LoadingScreen/>
  switch (route.path) {
    case '/explore': return <ExplorePage/>
    case '/create': return <CreatePostPage/>
    case '/notifications': return <PlaceholderPage route="/notifications" title="Notifications" description="Updates about your network and conversations will appear here." icon={faBell}/>
    case '/profile': return <ProfilePage profile={profile}/>
    case '/profile/edit': return <EditProfilePage profile={profile} onUpdate={updateProfile} onLogout={logout}/>
    case '/profile/sessions': return <SessionsPage onCurrentRevoked={clearSession}/>
    default: return <FeedPage/>
  }
}
