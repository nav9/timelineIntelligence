<script lang="ts">
  import { link, push } from 'svelte-spa-router'
  import { api, ApiRequestError } from '../lib/api'
  import { validatePassword, isValidEmail } from '../lib/password'
  import PasswordInput from '../components/PasswordInput.svelte'
  import PasswordStrength from '../components/PasswordStrength.svelte'

  let name = ''
  let email = ''
  let password = ''
  let error = ''
  let serverFeedback: string[] = []
  let loading = false
  let success = false

  $: pwCheck = validatePassword(password, name, email)

  async function onSubmit(e: Event) {
    e.preventDefault()
    error = ''
    serverFeedback = []

    if (!name.trim()) {
      error = 'Name is required'
      return
    }
    if (!isValidEmail(email.trim())) {
      error = 'Please enter a valid email address'
      return
    }
    if (!pwCheck.valid) {
      error = 'Password does not meet requirements'
      return
    }

    loading = true
    try {
      await api.register(name.trim(), email.trim(), password)
      success = true
      // Brief success, then go to login.
      setTimeout(() => push('/login'), 1200)
    } catch (err) {
      if (err instanceof ApiRequestError) {
        error = err.message
        serverFeedback = err.feedback
      } else {
        error = 'Registration failed. Please try again.'
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
      <h1>Create account</h1>
      <p>Register to use the Event &amp; Timeline Intelligence Platform</p>
    </div>

    {#if success}
      <div class="alert alert-success" role="status">Registration successful. Redirecting to sign in…</div>
    {:else}
      {#if error}
        <div class="alert alert-error" role="alert">{error}</div>
      {/if}
      {#if serverFeedback.length}
        <ul class="feedback-list">
          {#each serverFeedback as item}
            <li>{item}</li>
          {/each}
        </ul>
      {/if}

      <form on:submit={onSubmit} novalidate>
        <div class="form-group">
          <label class="form-label" for="reg-name">Name</label>
          <input
            id="reg-name"
            class="form-input"
            type="text"
            autocomplete="name"
            bind:value={name}
            required
            disabled={loading}
            placeholder="Your name"
          />
        </div>

        <div class="form-group">
          <label class="form-label" for="reg-email">Email</label>
          <input
            id="reg-email"
            class="form-input"
            type="email"
            autocomplete="email"
            bind:value={email}
            required
            disabled={loading}
            placeholder="you@example.com"
          />
        </div>

        <div class="form-group">
          <label class="form-label" for="reg-password">Password</label>
          <PasswordInput
            id="reg-password"
            bind:value={password}
            autocomplete="new-password"
            disabled={loading}
          />
          {#if password.length > 0}
            <PasswordStrength strength={pwCheck.strength} feedback={pwCheck.feedback} />
          {/if}
          <p class="form-hint">Use a long passphrase. Minimum 10 characters.</p>
        </div>

        <button
          type="submit"
          class="btn btn-primary btn-full"
          class:btn-loading={loading}
          disabled={loading || !name || !email || !password || !pwCheck.valid}
        >
          {loading ? 'Creating account…' : 'Register'}
        </button>
      </form>
    {/if}

    <div class="auth-footer">
      Already have an account?
      <a href="/login" use:link>Sign in</a>
    </div>
  </div>
</div>
