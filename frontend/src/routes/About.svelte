<div class="container main-content">
  <section class="page">
    <h1>About</h1>
    <p class="app-name">General Purpose Event &amp; Timeline Intelligence Platform</p>
    <p class="text-secondary" style="margin-top: 0.75rem;">
      Early development version. This release provides authentication, sessions, and a basic application shell.
    </p>
    <dl class="meta">
      <div><dt>Version</dt><dd>0.1.0</dd></div>
      <div><dt>Stack</dt><dd>Go / Gin · Svelte · SQLite</dd></div>
      <div><dt>Auth</dt><dd>Server-side sessions · Argon2id</dd></div>
    </dl>
  </section>
</div>

<style>
  .page { max-width: 560px; margin: 0 auto; }
  .page h1 { margin-bottom: var(--sp-3); font-size: var(--text-xl); }
  .app-name {
    font-size: var(--text-lg);
    font-weight: 600;
    color: var(--accent);
  }
  .meta {
    margin-top: var(--sp-6);
    display: flex;
    flex-direction: column;
    gap: var(--sp-3);
  }
  .meta > div {
    display: flex;
    gap: var(--sp-4);
    font-size: var(--text-sm);
    border-bottom: 1px solid var(--border);
    padding-bottom: var(--sp-2);
  }
  .meta dt {
    width: 5rem;
    flex-shrink: 0;
    color: var(--text-muted);
  }
  .meta dd {
    color: var(--text-secondary);
  }
</style>
