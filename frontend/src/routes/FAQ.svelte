<div class="container main-content">
  <section class="page">
    <h1>FAQ</h1>
    <dl class="faq">
      <dt>What is this application?</dt>
      <dd>An early foundation for a General Purpose Event &amp; Timeline Intelligence Platform. Only authentication and basic navigation are implemented so far.</dd>

      <dt>Is my password stored in plaintext?</dt>
      <dd>No. Passwords are hashed with Argon2id on the server. Plaintext passwords are never stored.</dd>

      <dt>What happens if I deactivate my account?</dt>
      <dd>Your account is marked deactivated, sessions are revoked, and login is blocked. Data is not immediately permanently deleted.</dd>
    </dl>
  </section>
</div>

<style>
  .page { max-width: 560px; margin: 0 auto; }
  .page h1 { margin-bottom: var(--sp-4); font-size: var(--text-xl); }
  .faq dt {
    font-size: var(--text-sm);
    font-weight: 600;
    color: var(--text-primary);
    margin-top: var(--sp-4);
  }
  .faq dd {
    font-size: var(--text-sm);
    color: var(--text-secondary);
    margin-top: var(--sp-1);
    padding-bottom: var(--sp-3);
    border-bottom: 1px solid var(--border);
  }
</style>
