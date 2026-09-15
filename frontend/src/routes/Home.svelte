<script lang="ts">
  import { currentUser } from '../lib/auth'
</script>

<div class="container main-content">
  <section class="landing">
    <h1>Login successful</h1>
    <p class="lead">
      Welcome{#if $currentUser}, {$currentUser.name}{/if}.
      This is the initial version of the
      <strong>Event &amp; Timeline Intelligence Platform</strong>.
    </p>
    <p class="text-secondary text-sm">
      Authentication is working. Event processing, timelines, and analysis will be added in later stages.
    </p>
  </section>
</div>

<style>
  .landing {
    max-width: 560px;
    margin: var(--sp-8) auto 0;
  }

  .landing h1 {
    margin-bottom: var(--sp-3);
    color: var(--success);
  }

  .lead {
    color: var(--text-primary);
    margin-bottom: var(--sp-4);
  }

  .lead strong {
    color: var(--accent);
    font-weight: 600;
  }
</style>
