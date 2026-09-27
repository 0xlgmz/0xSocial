import { faArrowLeft } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import type { ReactNode } from 'react'
import { navigate } from '../../lib/routes'
import type { Route } from '../../lib/routes'
import { BrandPanel } from '../brand/BrandPanel'
import { SiteFooter } from './SiteFooter'

export function PublicLayout({ children, back }: { children: ReactNode; back?: Route }) {
  return (
    <main className="min-h-screen bg-base-200 lg:flex">
      <BrandPanel/>
      <section className="flex min-h-screen flex-1 items-center justify-center p-5 sm:p-10">
        <div className="w-full max-w-md">
          {back && <button className="btn btn-ghost btn-sm mb-4 -ml-3 text-base-content/55" onClick={() => navigate(back)}><FontAwesomeIcon icon={faArrowLeft}/>Back</button>}
          {children}<SiteFooter/>
        </div>
      </section>
    </main>
  )
}
