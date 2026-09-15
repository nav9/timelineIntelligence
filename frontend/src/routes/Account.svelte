<script lang="ts">
  import { push } from 'svelte-spa-router'
  import { api, ApiRequestError } from '../lib/api'
  import { auth } from '../lib/auth'

  let confirmed = false
  let loading = false
  let error = ''
  let done = false

  async function deactivate() {
    if (!confirmed) return
    error = ''
    loading = true
    try {
      await api.deactivateAccount()
      done = true
      auth.clear()
      setTimeout(() => push('/login'), 2000)
    } catch (err) {
      if (err instanceof ApiRequestError) {
        error = err.message
      } else {
        error = 'Account deactivation failed.'
      }
    } finally {
      loading = false
    }
  }
</script>

<div class="container main-content">
  <section class="page">
    <h1>Unregister / Deactivate Account</h1>

    {#if done}
      <div class="alert alert-success" role="status">
        Your account has been deactivated. Redirecting to sign in…
      </div>
    {:else}
      <div class="alert alert-warning" role="note">
        <p>
          Deactivating your account will immediately revoke all active sessions and prevent future login.
          Your account record and associated data are <strong>not</strong> permanently deleted at this stage.
        </p>
        <p style="margin-top: 0.75rem;">
          Data may be retained for applicable legal, regulatory, contractual, or operational retention
          requirements. Retention periods are determined by applicable requirements and are not
          represented here as a fixed duration.
        </p>
      </div>

      {#if error}
        <div class="alert alert-error" role="alert">{error}</div>
      {/if}

      <label class="confirm">
        <input type="checkbox" bind:checked={confirmed} disabled={loading} />
        <span>I understand that my account will be deactivated and that data may be retained as described above.</span>
      </label>

      <div class="actions">
        <button
          type="button"
          class="btn btn-danger"
          disabled={!confirmed || loading}
          class:btn-loading={loading}
          on:click={deactivate}
        >
          {loading ? 'Deactivating…' : 'Deactivate Account'}
        </button>
        <button type="button" class="btn btn-secondary" disabled={loading} on:click={() => push('/')}>
          Cancel
        </button>
      </div>
    {/if}
  </section>
</div>

<style>
  .page {
    max-width: 560px;
    margin: 0 auto;
  }

  .page h1 {
    margin-bottom: var(--sp-4);
    font-size: var(--text-xl);
  }

  .confirm {
    display: flex;
    gap: var(--sp-3);
    align-items: flex-start;
    margin: var(--sp-5) 0;
    font-size: var(--text-sm);
    color: var(--text-secondary);
    cursor: pointer;
  }

  .confirm input {
    margin-top: 0.2rem;
    accent-color: var(--accent);
  }

  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: var(--sp-3);
  }
</style>
