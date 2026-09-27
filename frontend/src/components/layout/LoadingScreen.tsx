import { Logo } from '../brand/Logo'

export function LoadingScreen() {
  return <main className="grid min-h-screen place-items-center bg-base-100"><div className="flex flex-col items-center gap-5"><Logo/><span className="loading loading-spinner loading-md text-primary"/><span className="text-xs text-base-content/45">Connecting to your network…</span></div></main>
}
