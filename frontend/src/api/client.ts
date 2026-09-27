export class ApiError extends Error {
  status: number
  retryAfter: number | null

  constructor(status: number, message: string, retryAfter: number | null = null) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.retryAfter = retryAfter
  }
}

const friendlyErrors: Record<string, Record<number, string>> = {
  '/api/auth/register': {
    400: 'Check your email address and password requirements.',
    429: 'Too many registration attempts. Please try again later.',
  },
  '/api/auth/login': {
    401: 'The email or password is incorrect.',
    403: 'Please verify your email before logging in.',
    429: 'Too many login attempts. Please try again later.',
  },
  '/api/auth/verify-email': {
    400: 'This verification link is invalid or has expired.',
  },
  '/api/auth/reset-password': {
    400: 'The reset link is invalid, expired, or the password is not valid.',
    429: 'Too many reset attempts. Please try again later.',
  },
}

type RequestOptions = {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  body?: Record<string, unknown>
  signal?: AbortSignal
  authenticated?: boolean
}

export const sessionExpiredEvent = '0xsocial:session-expired'

export async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  let response: Response
  try {
    response = await fetch(path, {
      method: options.method ?? 'POST',
      credentials: 'include',
      signal: options.signal,
      headers: options.body ? { 'Content-Type': 'application/json' } : undefined,
      body: options.body ? JSON.stringify(options.body) : undefined,
    })
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') throw error
    throw new ApiError(0, 'We could not connect to 0xSocial. Check your connection and try again.')
  }

  if (!response.ok) {
    const backendMessage = (await response.text()).trim()
    const retryHeader = response.headers.get('retry-after')
    const retryAfter = retryHeader ? Number.parseInt(retryHeader, 10) : null
    const error = new ApiError(
      response.status,
      friendlyErrors[path]?.[response.status] || backendMessage || 'Something went wrong. Please try again.',
      Number.isFinite(retryAfter) ? retryAfter : null,
    )
    if (response.status === 401 && options.authenticated) window.dispatchEvent(new Event(sessionExpiredEvent))
    throw error
  }

  if (response.status === 204 || response.headers.get('content-length') === '0') return undefined as T
  const contentType = response.headers.get('content-type') ?? ''
  return contentType.includes('application/json') ? response.json() as Promise<T> : undefined as T
}
