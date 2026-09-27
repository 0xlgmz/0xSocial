import { faArrowLeft, faCompass } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { Logo } from '../components/brand/Logo'
import { navigate } from '../lib/routes'

export function PageNotFound() {
  return (
    <main className="min-h-screen bg-base-200">
      <header className="border-b border-base-300 bg-base-100"><div className="mx-auto flex h-16 max-w-3xl items-center px-5"><Logo/></div></header>
      <section className="mx-auto grid min-h-[calc(100vh-4rem)] max-w-lg place-items-center p-6 text-center">
        <div>
          <div className="mx-auto grid size-20 place-items-center rounded-full bg-primary/10 text-3xl text-primary"><FontAwesomeIcon icon={faCompass}/></div>
          <p className="mt-6 font-mono text-xs uppercase tracking-[.2em] text-primary">404</p>
          <h1 className="mt-2 text-3xl font-semibold tracking-[-.035em]">Page not found</h1>
          <p className="mt-3 text-sm leading-6 text-base-content/50">The page you’re looking for doesn’t exist or may have moved.</p>
          <button className="btn btn-primary mt-6" onClick={() => navigate('/')}><FontAwesomeIcon icon={faArrowLeft}/>Back to 0xSocial</button>
        </div>
      </section>
    </main>
  )
}
