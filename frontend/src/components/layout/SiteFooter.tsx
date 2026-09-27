import { navigate } from '../../lib/routes'
import type { LegalRoute } from '../../lib/routes'
import { useMonetization } from '../../features/monetization/useMonetization'

const links: Array<{ route: LegalRoute; label: string }> = [
  { route: '/privacy', label: 'Privacy' },
  { route: '/cookies', label: 'Cookies' },
  { route: '/terms', label: 'Terms' },
  { route: '/community-guidelines', label: 'Guidelines' },
  { route: '/contact', label: 'Contact' },
]

export function SiteFooter() {
  const { openPrivacyChoices } = useMonetization()

  function privacyChoices() {
    if (!openPrivacyChoices()) navigate('/privacy-choices')
  }

  return (
    <footer className="flex flex-wrap items-center justify-center gap-x-4 gap-y-2 px-4 py-6 text-[.7rem] text-base-content/45" aria-label="Site information">
      {links.map(link => <button key={link.route} className="link link-hover" onClick={() => navigate(link.route)}>{link.label}</button>)}
      <button className="link link-hover" onClick={privacyChoices}>Privacy choices</button>
      <span>© 2026 0xSocial</span>
    </footer>
  )
}
