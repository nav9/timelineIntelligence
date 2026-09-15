<div class="container main-content">
  <section class="page">
    <h1>Settings</h1>
    <p class="text-secondary">
      Settings placeholder. Future options may include theme, font size, row spacing, and display preferences.
    </p>
    <ul class="placeholder-list">
      <li>Theme <span class="badge badge-blue">coming later</span></li>
      <li>Font size <span class="badge badge-blue">coming later</span></li>
      <li>Row spacing <span class="badge badge-blue">coming later</span></li>
      <li>Display preferences <span class="badge badge-blue">coming later</span></li>
    </ul>
  </section>
</div>

<style>
  .page { max-width: 560px; margin: 0 auto; }
  .page h1 { margin-bottom: var(--sp-3); font-size: var(--text-xl); }
  .placeholder-list {
    list-style: none;
    margin-top: var(--sp-5);
    display: flex;
    flex-direction: column;
    gap: var(--sp-3);
  }
  .placeholder-list li {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--sp-3) 0;
    border-bottom: 1px solid var(--border);
    font-size: var(--text-sm);
    color: var(--text-secondary);
  }
</style>
