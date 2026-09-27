import { faEye, faEyeSlash } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { useState } from 'react'
import type { ReactNode } from 'react'

type PasswordFieldProps = {
  password: string
  onChange: (value: string) => void
  label?: string
  autoComplete?: 'new-password' | 'current-password'
  labelAction?: ReactNode
}

export function PasswordField({ password, onChange, label = 'Password', autoComplete = 'new-password', labelAction }: PasswordFieldProps) {
  const [visible, setVisible] = useState(false)
  return (
    <fieldset className="fieldset">
      <div className="flex items-center justify-between"><legend className="fieldset-legend">{label}</legend>{labelAction}</div>
      <div className="join w-full">
        <input className="input join-item w-full" type={visible ? 'text' : 'password'} autoComplete={autoComplete} required minLength={8} value={password} onChange={event => onChange(event.target.value)} placeholder="8+ characters"/>
        <button type="button" className="btn btn-square join-item border-base-300 bg-base-100" onClick={() => setVisible(value => !value)} aria-label={visible ? 'Hide password' : 'Show password'}>
          <FontAwesomeIcon icon={visible ? faEyeSlash : faEye}/>
        </button>
      </div>
    </fieldset>
  )
}
