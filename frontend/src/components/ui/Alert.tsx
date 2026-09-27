import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faCircleCheck, faShieldHalved, faTriangleExclamation } from '@fortawesome/free-solid-svg-icons'
import type { ReactNode } from 'react'

export function Alert({ type, children }: { type: 'error' | 'success' | 'info'; children: ReactNode }) {
  const icon = type === 'error'
    ? faTriangleExclamation
    : type === 'success' ? faCircleCheck : faShieldHalved
  return (
    <div role={type === 'error' ? 'alert' : 'status'} className={`alert alert-${type} items-start py-3 text-xs`}>
      <FontAwesomeIcon icon={icon} className="mt-0.5"/><span>{children}</span>
    </div>
  )
}
