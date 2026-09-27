import { faBell, faHouse, faMagnifyingGlass, faPlus, faUser } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import type { IconDefinition } from '@fortawesome/fontawesome-svg-core'
import type { ReactNode } from 'react'
import { useAuth } from '../../features/auth/useAuth'
import { navigate } from '../../lib/routes'
import type { PrivateRoute, Route } from '../../lib/routes'
import { ProfileAvatar } from '../user/ProfileAvatar'
import { SiteFooter } from './SiteFooter'

const navigation: { route: PrivateRoute; label: string; icon: IconDefinition }[] = [
  { route: '/feed', label: 'Home', icon: faHouse },
  { route: '/explore', label: 'Explore', icon: faMagnifyingGlass },
  { route: '/create', label: 'Create', icon: faPlus },
  { route: '/notifications', label: 'Notifications', icon: faBell },
  { route: '/profile', label: 'You', icon: faUser },
]

function BottomNavigation({ route }: { route: Route | null }) {
  const { profile } = useAuth()
  return (
    <nav className="fixed inset-x-4 bottom-[max(1rem,env(safe-area-inset-bottom))] z-30 mx-auto grid h-[4.5rem] max-w-xl grid-cols-5 items-center rounded-[1.6rem] border border-base-300/80 bg-base-100/95 px-2 shadow-[0_12px_35px_rgba(20,38,29,0.16)] backdrop-blur-xl" aria-label="Primary navigation">
      {navigation.map(item => {
        const active = route === item.route || (item.route === '/profile' && (route === '/profile/edit' || route === '/profile/sessions'))
        return (
          <button
            key={item.route}
            type="button"
            className={`group grid h-14 place-items-center rounded-2xl text-xl transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary ${active ? 'text-primary' : 'text-base-content/45 hover:bg-base-200 hover:text-base-content'}`}
            onClick={() => navigate(item.route)}
            aria-label={item.label}
            aria-current={active ? 'page' : undefined}
          >
            {item.route === '/profile' && profile ? (
              <span className={active ? 'rounded-full ring-2 ring-primary ring-offset-2 ring-offset-base-100' : 'rounded-full'}><ProfileAvatar displayName={profile.displayName} fallbackText={profile.handle || profile.email} avatarUrl={profile.avatarUrl} size="sm"/></span>
            ) : (
              <span className={item.route === '/create' ? 'grid size-10 place-items-center rounded-full bg-primary text-primary-content shadow-sm transition-transform group-hover:scale-105' : ''}><FontAwesomeIcon icon={item.icon}/></span>
            )}
          </button>
        )
      })}
    </nav>
  )
}

export function AuthenticatedLayout({ title, route, children }: { title: string; route: Route | null; children: ReactNode }) {
  return (
    <div className="min-h-screen bg-base-200 pb-28">
      <main className="mx-auto max-w-2xl p-4 sm:p-6" aria-label={title}>{children}<SiteFooter/></main>
      <BottomNavigation route={route}/>
    </div>
  )
}
