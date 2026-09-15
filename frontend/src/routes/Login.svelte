<script lang="ts">
  import { link, push } from 'svelte-spa-router'
  import { api, ApiRequestError } from '../lib/api'
  import { auth } from '../lib/auth'
  import PasswordInput from '../components/PasswordInput.svelte'

  let email = ''
  let password = ''
  let error = ''
  let loading = false

  async function onSubmit(e: Event) {
    e.preventDefault()
    error = ''
    loading = true
    try {
      const res = await api.login(email.trim(), password)
      auth.setUser(res.user)
      push('/')
    } catch (err) {
      if (err instanceof ApiRequestError) {
        error = err.message
      } else {
        error = 'Login failed. Please try again.'
      }
    } finally {
      loading = false
    }
  }
</script>

<div class="auth-page">
  <div class="auth-card">
    <div class="auth-header">
      <div class="auth-logo">
        <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#3b82f6" stroke-width="2" aria-hidden="true">
          <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/>
        </svg>
        <span>Timeline Intelligence</span>
      </div>
      <h1>Sign in</h1>
      <p>Access your Event &amp; Timeline Intelligence account</p>
    </div>

    {#if error}
      <div class="alert alert-error" role="alert">{error}</div>
    {/if}

    <form on:submit={onSubmit} novalidate>
      <div class="form-group">
        <label class="form-label" for="login-email">Email</label>
        <input
          id="login-email"
          class="form-input"
          type="email"
          autocomplete="username"
          bind:value={email}
          required
          disabled={loading}
          placeholder="you@example.com"
        />
      </div>

      <div class="form-group">
        <label class="form-label" for="login-password">Password</label>
        <PasswordInput
          id="login-password"
          bind:value={password}
          autocomplete="current-password"
          disabled={loading}
        />
      </div>

      <button type="submit" class="btn btn-primary btn-full" class:btn-loading={loading} disabled={loading || !email || !password}>
        {loading ? 'Signing in…' : 'Sign in'}
      </button>
    </form>

    <div class="auth-footer">
      No account?
      <a href="/register" use:link>Register</a>
    </div>
  </div>
</div>
