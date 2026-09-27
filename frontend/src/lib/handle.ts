export function normalizeHandle(handle: string) {
  return handle.trim().toLowerCase()
}

export function validateHandle(handle: string) {
  const normalized = normalizeHandle(handle)
  if (!normalized) return 'Choose a handle.'
  if (normalized.length < 3) return 'Handle must be at least 3 characters.'
  if (normalized.length > 30) return 'Handle must be no more than 30 characters.'
  if (!/^[a-z]/.test(normalized)) return 'Handle must start with a letter.'
  if (!/^[a-z][a-z0-9_]*$/.test(normalized)) {
    return 'Use only lowercase letters, numbers, and underscores.'
  }
  return ''
}
