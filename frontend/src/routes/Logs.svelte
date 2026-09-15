<script lang="ts">
  import { currentUser } from '../lib/auth'

  $: isAdmin = $currentUser?.role === 'admin'
</script>

<div class="container main-content">
  <section class="page">
    <h1>Logs</h1>

    {#if isAdmin}
      <p class="text-secondary">
        Administrator log access is authorized for your account. Full audit-log UI is not yet implemented.
      </p>
      <p class="text-sm text-muted" style="margin-top: 0.75rem;">
        The backend exposes <code>GET /api/admin/logs</code> behind admin authorization.
        Remaining work: pagination UI, filtering, and retention controls.
      </p>
    {:else}
      <div class="alert alert-warning" role="status">
        Sensitive logs are restricted to administrators.
        Your account does not have administrator access.
      </div>
      <p class="text-sm text-muted" style="margin-top: 1rem;">
        Authorization boundary is enforced on the backend. Ordinary authenticated users cannot retrieve audit logs.
      </p>
    {/if}
  </section>
</div>

<style>
  .page { max-width: 560px; margin: 0 auto; }
  .page h1 { margin-bottom: var(--sp-3); font-size: var(--text-xl); }
</style>
