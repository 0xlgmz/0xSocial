import { faCheck } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { getPasswordChecks } from '../../lib/password'

export function PasswordChecklist({ password }: { password: string }) {
  const isPassphrase = password.length >= 15
  return (
    <div className="rounded-box bg-base-200 p-3">
      <p className="mb-2 text-[.65rem] font-semibold text-base-content/60">{isPassphrase ? 'Passphrase requirements' : 'Password requirements'}</p>
      <div className="grid grid-cols-1 gap-1.5 sm:grid-cols-2">
        {getPasswordChecks(password).map(({ label, valid }) => (
          <span key={label} className={`flex items-center gap-2 text-[.65rem] ${valid ? 'text-success' : 'text-base-content/35'}`}>
            <span className={`grid size-4 place-items-center rounded-full ${valid ? 'bg-success/15' : 'bg-base-300'}`}>{valid && <FontAwesomeIcon icon={faCheck} className="text-[.5rem]"/>}</span>{label}
          </span>
        ))}
      </div>
    </div>
  )
}
