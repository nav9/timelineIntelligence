/**
 * HTTP client for the backend API.
 * Handles CSRF tokens and credentials (session cookies).
 * Security decisions remain on the backend — this is presentation plumbing only.
 */

const CSRF_COOKIE = 'tl_csrf'
const CSRF_HEADER = 'X-CSRF-Token'

export interface ApiError {
  error: string
  feedback?: string[]
}

export class ApiRequestError extends Error {
  status: number
  feedback: string[]

  constructor(status: number, message: string, feedback: string[] = []) {
    super(message)
    this.name = 'ApiRequestError'
    this.status = status
    this.feedback = feedback
  }
}

/** Read a cookie value by name. */
function getCookie(name: string): string | null {
  const match = document.cookie.match(new RegExp('(?:^|; )' + name.replace(/([.$?*|{}()[\]\\/+^])/g, '\\$1') + '=([^;]*)'))
  return match ? decodeURIComponent(match[1]) : null
}

/**
 * Ensure a CSRF cookie exists by hitting a safe GET endpoint first.
 * The backend issues the CSRF cookie on GET requests.
 */
export async function ensureCsrf(): Promise<void> {
  if (!getCookie(CSRF_COOKIE)) {
    await fetch('/api/health', { credentials: 'include' })
  }
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = {
    Accept: 'application/json',
  }

  if (body !== undefined) {
    headers['Content-Type'] = 'application/json'
  }

  // CSRF required for state-changing methods.
  if (method !== 'GET' && method !== 'HEAD') {
    await ensureCsrf()
    const csrf = getCookie(CSRF_COOKIE)
    if (csrf) {
      headers[CSRF_HEADER] = csrf
    }
  }

  const res = await fetch(path, {
    method,
    headers,
    credentials: 'include',
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })

  let data: unknown = null
  const contentType = res.headers.get('content-type') || ''
  if (contentType.includes('application/json')) {
    data = await res.json()
  }

  if (!res.ok) {
    const err = data as ApiError | null
    throw new ApiRequestError(
      res.status,
      err?.error || `Request failed (${res.status})`,
      err?.feedback || [],
    )
  }

  return data as T
}

export interface SafeUser {
  id: number
  name: string
  email: string
  status: string
  role: string
  created_at: string
}

export interface AuthResponse {
  message: string
  user: SafeUser
}

export interface MeResponse {
  user: SafeUser
}

export interface HealthResponse {
  status: string
  timestamp: string
  components: {
    database: string
    api: string
  }
}

export const api = {
  health: () => request<HealthResponse>('GET', '/api/health'),

  register: (name: string, email: string, password: string) =>
    request<AuthResponse>('POST', '/api/auth/register', { name, email, password }),

  login: (email: string, password: string) =>
    request<AuthResponse>('POST', '/api/auth/login', { email, password }),

  logout: () => request<{ message: string }>('POST', '/api/auth/logout'),

  me: () => request<MeResponse>('GET', '/api/auth/me'),

  deactivateAccount: () =>
    request<{ message: string }>('DELETE', '/api/auth/account'),
}
