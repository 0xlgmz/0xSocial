import type { ResolvedRoute, Route } from './routes'

const routeTitles: Record<Route, string | null> = {
  '/': null,
  '/register': 'Create Account',
  '/check-email': 'Check Your Email',
  '/verify-email': 'Verify Email',
  '/login': 'Log In',
  '/forgot-password': 'Forgot Password',
  '/reset-password': 'Reset Password',
  '/privacy': 'Privacy Policy',
  '/cookies': 'Cookie Policy',
  '/terms': 'Terms of Service',
  '/contact': 'Contact',
  '/community-guidelines': 'Community Guidelines',
  '/privacy-choices': 'Privacy Choices',
  '/feed': 'Feed',
  '/explore': 'Explore',
  '/create': 'Create Post',
  '/notifications': 'Notifications',
  '/profile': 'Profile',
  '/profile/edit': 'Edit Profile',
  '/profile/sessions': 'Active Sessions',
}

export function getDocumentTitle(route: ResolvedRoute) {
  if (route.kind === 'profile') return `@${route.handle} | 0xSocial`
  if (route.kind === 'not-found') return 'Page Not Found | 0xSocial'

  const pageTitle = routeTitles[route.path]
  return pageTitle ? `${pageTitle} | 0xSocial` : '0xSocial'
}
