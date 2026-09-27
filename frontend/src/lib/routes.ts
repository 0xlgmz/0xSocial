import { normalizeHandle, validateHandle } from './handle'

export type PublicRoute =
  | '/'
  | '/register'
  | '/check-email'
  | '/verify-email'
  | '/login'
  | '/forgot-password'
  | '/reset-password'
  | LegalRoute

export type LegalRoute =
  | '/privacy'
  | '/cookies'
  | '/terms'
  | '/contact'
  | '/community-guidelines'
  | '/privacy-choices'

export type PrivateRoute =
  | '/feed'
  | '/explore'
  | '/create'
  | '/notifications'
  | '/profile'
  | '/profile/edit'
  | '/profile/sessions'

export type Route = PublicRoute | PrivateRoute

export type ResolvedRoute =
  | { kind: 'static'; path: Route }
  | { kind: 'profile'; handle: string }
  | { kind: 'not-found' }

export const publicRoutes = new Set<PublicRoute>([
  '/', '/register', '/check-email', '/verify-email', '/login',
  '/forgot-password', '/reset-password', '/privacy', '/cookies', '/terms',
  '/contact', '/community-guidelines', '/privacy-choices',
])

export const legalRoutes = new Set<LegalRoute>([
  '/privacy', '/cookies', '/terms', '/contact', '/community-guidelines', '/privacy-choices',
])

export const privateRoutes = new Set<PrivateRoute>([
  '/feed', '/explore', '/create', '/notifications', '/profile', '/profile/edit', '/profile/sessions',
])

export function resolveRoute(pathname: string): ResolvedRoute {
  const profileMatch = pathname.match(/^\/@([a-z][a-z0-9_]{2,29})$/)
  if (profileMatch) return { kind: 'profile', handle: profileMatch[1] }

  if (publicRoutes.has(pathname as PublicRoute) || privateRoutes.has(pathname as PrivateRoute)) {
    return { kind: 'static', path: pathname as Route }
  }

  return { kind: 'not-found' }
}

export function getCurrentRoute() {
  return resolveRoute(window.location.pathname)
}

export function navigatePath(path: string) {
  window.history.pushState({}, '', path)
  window.dispatchEvent(new PopStateEvent('popstate'))
}

export function replacePath(path: string) {
  window.history.replaceState({}, '', path)
  window.dispatchEvent(new PopStateEvent('popstate'))
}

export function navigate(route: Route, params?: Record<string, string>) {
  const query = params ? `?${new URLSearchParams(params)}` : ''
  navigatePath(`${route}${query}`)
}

export function redirect(route: Route, params?: Record<string, string>) {
  const query = params ? `?${new URLSearchParams(params)}` : ''
  replacePath(`${route}${query}`)
}

export function navigateToProfile(handle: string) {
  const normalized = normalizeHandle(handle)
  if (validateHandle(normalized)) return
  navigatePath(`/@${encodeURIComponent(normalized)}`)
}

export function getQueryParam(name: string) {
  return new URLSearchParams(window.location.search).get(name) ?? ''
}
