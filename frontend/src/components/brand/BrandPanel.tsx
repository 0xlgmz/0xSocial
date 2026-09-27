import { Logo } from './Logo'

export function BrandPanel() {
  return (
    <aside className="brand-panel hidden min-h-screen overflow-hidden bg-neutral text-neutral-content lg:flex lg:w-1/2 lg:flex-col lg:justify-between">
      <Logo light/>
      <div className="relative z-10 max-w-lg">
        <p className="mb-4 font-mono text-xs uppercase tracking-[.28em] text-primary-content/65">A quieter place to connect</p>
        <h2 className="text-5xl font-semibold leading-[1.08] tracking-[-.045em]">Your people.<br/>Your feed.</h2>
        <p className="mt-6 max-w-sm text-sm leading-7 text-neutral-content/60">Share what matters, follow conversations you care about, and stay close to your network.</p>
      </div>
      <p className="relative z-10 text-xs text-neutral-content/35">© 2026 0xSocial</p>
      <div className="brand-orb brand-orb-one"/><div className="brand-orb brand-orb-two"/>
    </aside>
  )
}
