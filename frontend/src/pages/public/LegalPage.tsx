import type { ReactNode } from 'react'
import { Logo } from '../../components/brand/Logo'
import { SiteFooter } from '../../components/layout/SiteFooter'
import { useMonetization } from '../../features/monetization/useMonetization'
import { navigate } from '../../lib/routes'
import type { LegalRoute } from '../../lib/routes'

const updated = 'September 27, 2026'

function Section({ title, children }: { title: string; children: ReactNode }) {
  return <section><h2 className="text-lg font-semibold">{title}</h2><div className="mt-2 space-y-3 text-sm leading-7 text-base-content/65">{children}</div></section>
}

const content: Record<Exclude<LegalRoute, '/privacy-choices'>, { title: string; intro: string; body: ReactNode }> = {
  '/privacy': {
    title: 'Privacy Policy',
    intro: 'How 0xSocial collects, uses, and shares information.',
    body: <>
      <Section title="Information we process"><p>We process account details, profile information, posts, uploaded media, reports, session and security records, and technical information needed to operate and protect the service.</p></Section>
      <Section title="Advertising and Google"><p>When advertising is enabled and your privacy choices permit it, Google and its advertising partners may use cookies, local storage, IP addresses, device or browser identifiers, and information about ad interactions to provide, measure, limit, or personalize ads. Google acts as a third-party processor or independent controller for some of these activities under its own terms.</p><p>0xSocial does not manually send your post text or profile fields to Google as ad-targeting parameters. Ads may still use contextual information from the page and coarse location information handled by Google.</p><p>Learn more in <a className="link" href="https://policies.google.com/technologies/partner-sites" target="_blank" rel="noreferrer">How Google uses information from sites or apps that use its services</a>.</p></Section>
      <Section title="Why we use information"><p>We use information to provide accounts and social features, secure the service, enforce our rules, respond to reports, improve reliability, comply with law, and—where enabled—fund the service with advertising.</p></Section>
      <Section title="Your choices"><p>You may change advertising consent through Privacy choices, update your account information, or contact us about access, correction, deletion, restriction, portability, or objection rights that apply to you.</p></Section>
      <Section title="Retention and sharing"><p>We keep information only as long as reasonably needed for the purposes above, legal obligations, disputes, and security. We share data with infrastructure, security, email, storage, moderation, analytics, and advertising providers only as needed for their services or when law requires it.</p></Section>
    </>,
  },
  '/cookies': {
    title: 'Cookie Policy',
    intro: 'How cookies and similar technologies are used on 0xSocial.',
    body: <>
      <Section title="Essential storage"><p>We use essential cookies and local storage for secure sessions, account protection, navigation, and preferences such as your theme. Disabling them may prevent core features from working.</p></Section>
      <Section title="Advertising storage"><p>When ads are enabled, Google and participating advertising partners may use cookies, local storage, and device identifiers for ad delivery, fraud prevention, frequency controls, measurement, and—only when permitted—personalization.</p></Section>
      <Section title="Control"><p>A Google-certified consent platform records and communicates applicable choices using industry consent strings. You can accept, reject, manage, or later withdraw choices through the persistent Privacy choices link. Browser controls can also delete or block storage, but may affect the service.</p></Section>
    </>,
  },
  '/terms': {
    title: 'Terms of Service',
    intro: 'The basic rules for using 0xSocial.',
    body: <>
      <Section title="Using the service"><p>You must provide accurate account information, keep your account secure, and use the service lawfully. You are responsible for content you post and for having the rights needed to share it.</p></Section>
      <Section title="Content and conduct"><p>You retain ownership of your content and grant 0xSocial a limited license to host, reproduce, and display it for operating the service. Content that violates the Community Guidelines may be restricted or removed, and serious or repeated violations may lead to account action.</p></Section>
      <Section title="Service availability"><p>The service is provided as available and may change or experience interruptions. To the extent permitted by law, 0xSocial disclaims implied warranties and is not liable for indirect or consequential loss.</p></Section>
    </>,
  },
  '/contact': {
    title: 'Contact',
    intro: 'Questions, privacy requests, and safety concerns.',
    body: <>
      <Section title="Get in touch"><p>Email <a className="link" href="mailto:support@0xsocial.app">support@0xsocial.app</a> for general help and <a className="link" href="mailto:privacy@0xsocial.app">privacy@0xsocial.app</a> for privacy requests. For a post that may violate our rules, use Report post so the relevant content identifier is included.</p></Section>
      <Section title="Urgent matters"><p>If someone is in immediate danger, contact local emergency services. Do not use 0xSocial reporting as a substitute for emergency help.</p></Section>
    </>,
  },
  '/community-guidelines': {
    title: 'Community Guidelines',
    intro: 'Help keep 0xSocial safe, useful, and human.',
    body: <>
      <Section title="Be respectful"><p>Do not harass, threaten, exploit, impersonate, or target people with hateful conduct. Do not promote violence, self-harm, or illegal activity.</p></Section>
      <Section title="Be authentic"><p>Do not spam, manipulate engagement, spread harmful deception, distribute malware, or use the service for fraud. Respect privacy, intellectual property, and consent.</p></Section>
      <Section title="Reporting"><p>Use Report post when content may break these rules. We confirm receipt without disclosing confidential moderation signals or promising a particular outcome. Misuse of reporting tools is itself prohibited.</p></Section>
    </>,
  },
}

export function LegalPage({ route }: { route: LegalRoute }) {
  const { openPrivacyChoices, consentStatus } = useMonetization()

  if (route === '/privacy-choices') {
    return <PageShell title="Privacy and cookie choices" intro="Review or change your advertising privacy choices."><Section title="Manage your choices"><p>0xSocial uses Google Privacy & Messaging, a Google-certified consent platform, to present controls where applicable and communicate your decision using IAB consent signals.</p><button className="btn btn-primary mt-2" onClick={() => openPrivacyChoices()} disabled={consentStatus === 'checking'}>Open privacy choices</button>{consentStatus === 'unavailable' && <p className="text-warning">Privacy controls are currently unavailable. Advertising remains disabled until the consent service is available.</p>}</Section><Section title="Other controls"><p>You can also clear site data in your browser. See the Cookie Policy for the kinds of storage used.</p></Section></PageShell>
  }

  const page = content[route]
  return <PageShell title={page.title} intro={page.intro}>{page.body}</PageShell>
}

function PageShell({ title, intro, children }: { title: string; intro: string; children: ReactNode }) {
  return <div className="min-h-screen bg-base-200"><header className="border-b border-base-300 bg-base-100"><div className="mx-auto flex max-w-3xl items-center justify-between p-4 sm:px-6"><Logo/><button className="btn btn-ghost btn-sm" onClick={() => navigate('/')}>Back to 0xSocial</button></div></header><main className="mx-auto max-w-3xl p-5 sm:p-8"><article className="card border border-base-300 bg-base-100 shadow-sm"><div className="card-body gap-8 p-6 sm:p-10"><header><p className="text-xs uppercase tracking-widest text-primary">0xSocial</p><h1 className="mt-2 text-3xl font-semibold tracking-tight">{title}</h1><p className="mt-3 text-sm leading-7 text-base-content/55">{intro}</p><p className="mt-2 text-xs text-base-content/40">Last updated {updated}</p></header>{children}</div></article></main><SiteFooter/></div>
}
