/**
 * Authentication store — holds the current user and loading state.
 * Session cookies are managed by the browser; this store only mirrors /api/auth/me.
 */

import { writable, derived, get } from 'svelte/store'
import { api, type SafeUser, ApiRequestError } from './api'

export type AuthStatus = 'unknown' | 'authenticated' | 'anonymous'

interface AuthState {
  status: AuthStatus
  user: SafeUser | null
  loading: boolean
  error: string | null
}

const initial: AuthState = {
  status: 'unknown',
  user: null,
  loading: true,
  error: null,
}

function createAuthStore() {
  const store = writable<AuthState>(initial)
  const { subscribe, set, update } = store

  return {
    subscribe,

    /** Fetch current session from the backend. */
    async refresh(): Promise<void> {
      update((s) => ({ ...s, loading: true, error: null }))
      try {
        const res = await api.me()
        set({ status: 'authenticated', user: res.user, loading: false, error: null })
      } catch (err) {
        if (err instanceof ApiRequestError && err.status === 401) {
          set({ status: 'anonymous', user: null, loading: false, error: null })
        } else {
          set({
            status: 'anonymous',
            user: null,
            loading: false,
            error: err instanceof Error ? err.message : 'Failed to check session',
          })
        }
      }
    },

    setUser(user: SafeUser): void {
      set({ status: 'authenticated', user, loading: false, error: null })
    },

    clear(): void {
      set({ status: 'anonymous', user: null, loading: false, error: null })
    },

    getUser(): SafeUser | null {
      return get(store).user
    },
  }
}

export const auth = createAuthStore()

export const isAuthenticated = derived(auth, ($a) => $a.status === 'authenticated')
export const currentUser = derived(auth, ($a) => $a.user)
