export type LoginRequest = {
  email: string
  password: string
}

export type Session = {
  email: string
  status: 'pending_verification' | 'active'
  expiresAt: string
}

export async function login(input: LoginRequest): Promise<void> {
  const response = await fetch('/api/auth/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include',
    body: JSON.stringify(input),
  })

  if (response.status === 401) {
    throw new Error('Invalid email or password')
  }

  if (!response.ok) {
    throw new Error('Unable to log in right now')
  }
}

export async function logout(): Promise<void> {
  const response = await fetch('/api/auth/logout', {
    method: 'POST',
    credentials: 'include',
  })

  if (!response.ok) {
    throw new Error('Unable to log out right now')
  }
}

export async function getSession(signal?: AbortSignal): Promise<Session | null> {
  const response = await fetch('/api/auth/me', {
    credentials: 'include',
    signal,
  })

  if (response.status === 401) {
    return null
  }

  if (!response.ok) {
    throw new Error('Unable to check your session')
  }

  return response.json() as Promise<Session>
}
