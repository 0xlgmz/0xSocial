import { faMoon, faSun } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { useEffect, useState } from 'react'
import { getPreferredTheme, saveTheme, themeChangeEvent } from '../../lib/theme'
import type { Theme } from '../../lib/theme'

const options: { value: Theme; label: string; icon: typeof faSun }[] = [
  { value: 'light', label: 'Light', icon: faSun },
  { value: 'dark', label: 'Dark', icon: faMoon },
]

export function ThemeSelector() {
  const [theme, setTheme] = useState<Theme>(getPreferredTheme)

  useEffect(() => {
    const syncTheme = (event: Event) => setTheme((event as CustomEvent<Theme>).detail)
    window.addEventListener(themeChangeEvent, syncTheme)
    return () => window.removeEventListener(themeChangeEvent, syncTheme)
  }, [])

  return (
    <fieldset>
      <legend className="sr-only">Color theme</legend>
      <div className="grid grid-cols-2 gap-2 rounded-box bg-base-200 p-1.5">
        {options.map(option => {
          const selected = theme === option.value
          return (
            <button
              key={option.value}
              type="button"
              className={`btn btn-sm border-0 ${selected ? 'bg-base-100 text-primary shadow-sm hover:bg-base-100' : 'btn-ghost text-base-content/55'}`}
              onClick={() => saveTheme(option.value)}
              aria-pressed={selected}
            >
              <FontAwesomeIcon icon={option.icon}/>{option.label}
            </button>
          )
        })}
      </div>
    </fieldset>
  )
}
