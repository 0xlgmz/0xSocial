import { faRotateRight, faTriangleExclamation } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'

export function ConnectionError({ onRetry }: { onRetry: () => void }) {
  return <main className="grid min-h-screen place-items-center bg-base-200 p-6"><div className="card w-full max-w-md border border-base-300 bg-base-100 shadow-sm"><div className="card-body items-center text-center"><div className="grid size-16 place-items-center rounded-full bg-error/10 text-error"><FontAwesomeIcon icon={faTriangleExclamation} className="text-2xl"/></div><h1 className="card-title mt-3 text-2xl">We can’t connect right now</h1><p className="text-sm leading-6 text-base-content/60">0xSocial could not reach the server. Check your connection and try again.</p><button className="btn btn-primary mt-4 w-full" onClick={onRetry}><FontAwesomeIcon icon={faRotateRight}/>Try again</button></div></div></main>
}
