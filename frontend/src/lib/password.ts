export function getPasswordChecks(password: string) {
  const isPassphrase = password.length >= 15
  return [
    { label: 'At least 8 characters', valid: password.length >= 8 },
    {
      label: 'No more than 1,024 bytes',
      valid: new TextEncoder().encode(password).length <= 1024,
    },
    {
      label: 'One uppercase letter',
      valid: isPassphrase || /[A-Z]/.test(password),
    },
    {
      label: 'One symbol',
      valid: isPassphrase || /[^A-Za-z0-9\s]/.test(password),
    },
  ]
}

export function validatePassword(password: string) {
  if (password.length < 8) return 'Password must be at least 8 characters.'
  if (new TextEncoder().encode(password).length > 1024) {
    return 'Password must be no more than 1,024 bytes.'
  }
  if (password.length < 15 && !/[A-Z]/.test(password)) {
    return 'Passwords under 15 characters need an uppercase letter.'
  }
  if (password.length < 15 && !/[^A-Za-z0-9\s]/.test(password)) {
    return 'Passwords under 15 characters need a symbol.'
  }
  return ''
}
