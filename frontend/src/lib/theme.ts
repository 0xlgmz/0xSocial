export const themes = ['light', 'dark'] as const
export type Theme = typeof themes[number]

const storageKey = 'theme'
export const themeChangeEvent = '0xsocial:theme-change'

function isTheme(value: string | null): value is Theme {
  return themes.includes(value as Theme)
}

export function getSavedTheme(): Theme | null {
  try {
    const savedTheme = window.localStorage.getItem(storageKey)
    return isTheme(savedTheme) ? savedTheme : null
  } catch {
    return null
  }
}

export function getSystemTheme(): Theme {
  return window.matchMedia?.('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

export function getPreferredTheme(): Theme {
  return getSavedTheme() ?? getSystemTheme()
}

export function applyTheme(theme: Theme) {
  document.documentElement.dataset.theme = theme
  window.dispatchEvent(new CustomEvent<Theme>(themeChangeEvent, { detail: theme }))
}

export function saveTheme(theme: Theme) {
  try {
    window.localStorage.setItem(storageKey, theme)
  } catch {
    // The selected theme still applies for this session when storage is unavailable.
  }
  applyTheme(theme)
}

export function initializeTheme() {
  applyTheme(getPreferredTheme())

  const colorScheme = window.matchMedia?.('(prefers-color-scheme: dark)')
  colorScheme?.addEventListener('change', event => {
    if (!getSavedTheme()) applyTheme(event.matches ? 'dark' : 'light')
  })

  window.addEventListener('storage', event => {
    if (event.key === storageKey) applyTheme(getPreferredTheme())
  })
}
