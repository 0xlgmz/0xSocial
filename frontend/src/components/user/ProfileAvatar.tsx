import { faUser } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'

type ProfileAvatarProps = {
  displayName: string
  fallbackText: string
  avatarUrl: string | null
  size?: 'sm' | 'lg'
}

function getInitials(displayName: string, fallbackText: string) {
  const name = displayName.trim()
  if (name) {
    return name
      .split(/\s+/)
      .slice(0, 2)
      .map(part => part[0])
      .join('')
      .toUpperCase()
  }
  return fallbackText.replace(/^@/, '').slice(0, 2).toUpperCase()
}

export function ProfileAvatar({ displayName, fallbackText, avatarUrl, size = 'lg' }: ProfileAvatarProps) {
  const sizeClass = size === 'lg' ? 'size-24 text-2xl' : 'size-10 text-sm'
  const label = displayName.trim() || fallbackText || 'User'
  const initials = getInitials(displayName, fallbackText)

  if (avatarUrl) {
    return (
      <div className="avatar">
        <div className={`${sizeClass} relative grid place-items-center overflow-hidden rounded-full bg-primary font-semibold text-primary-content`}>
          <span aria-hidden="true">{initials || <FontAwesomeIcon icon={faUser}/>}</span>
          <img className="absolute inset-0 size-full object-cover" src={avatarUrl} alt={`${label}'s avatar`} referrerPolicy="no-referrer" onError={event => { event.currentTarget.style.display = 'none' }}/>
        </div>
      </div>
    )
  }

  return (
    <div className="avatar placeholder">
      <div className={`${sizeClass} grid place-items-center rounded-full bg-primary font-semibold text-primary-content`} aria-label={`${label}'s avatar`}>
        {initials || <FontAwesomeIcon icon={faUser}/>} 
      </div>
    </div>
  )
}
