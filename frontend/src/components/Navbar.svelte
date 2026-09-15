<script lang="ts">
  import { link, location, push } from 'svelte-spa-router'
  import { auth, currentUser } from '../lib/auth'
  import { api } from '../lib/api'
  import AccountMenu from './AccountMenu.svelte'

  let menuOpen = false

  function toggleMenu() {
    menuOpen = !menuOpen
  }

  function closeMenu() {
    menuOpen = false
  }

  async function handleLogout() {
    closeMenu()
    try {
      await api.logout()
    } catch {
      // Session may already be invalid — still clear local state.
    }
    auth.clear()
    push('/login')
  }

  function handleNavigate(path: string) {
    closeMenu()
    push(path)
  }

  // Close menu on route change.
  $: if ($location) {
    menuOpen = false
  }
</script>

<nav class="navbar" aria-label="Main navigation">
  <div class="navbar-inner">
    <a href="/" use:link class="brand" aria-label="Home">
      <svg class="brand-icon" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
        <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/>
      </svg>
      <span class="brand-text">Timeline Intelligence</span>
    </a>

    <div class="navbar-right">
      {#if $currentUser}
        <span class="user-name text-sm text-secondary" title={$currentUser.email}>
          {$currentUser.name}
        </span>
        <div class="account-wrap">
          <button
            type="button"
            class="account-btn"
            on:click={toggleMenu}
            aria-haspopup="true"
            aria-expanded={menuOpen}
            aria-label="Account menu"
          >
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
              <circle cx="12" cy="7" r="4"/>
            </svg>
          </button>
          {#if menuOpen}
            <AccountMenu
              on:navigate={(e) => handleNavigate(e.detail)}
              on:logout={handleLogout}
              on:close={closeMenu}
            />
          {/if}
        </div>
      {/if}
    </div>
  </div>
</nav>

<!-- Backdrop to close menu on outside click -->
{#if menuOpen}
  <button type="button" class="menu-backdrop" aria-label="Close menu" on:click={closeMenu}></button>
{/if}

<style>
  .navbar {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    height: var(--navbar-h);
    background: var(--bg-secondary);
    border-bottom: 1px solid var(--border);
    z-index: 100;
  }

  .navbar-inner {
    height: 100%;
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 var(--sp-4);
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .brand {
    display: flex;
    align-items: center;
    gap: var(--sp-2);
    color: var(--text-primary);
    text-decoration: none;
    font-size: var(--text-sm);
    font-weight: 600;
  }

  .brand:hover {
    color: var(--accent);
  }

  .brand-icon {
    color: var(--accent);
    flex-shrink: 0;
  }

  .navbar-right {
    display: flex;
    align-items: center;
    gap: var(--sp-3);
  }

  .user-name {
    max-width: 140px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .account-wrap {
    position: relative;
  }

  .account-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 2rem;
    height: 2rem;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--bg-tertiary);
    color: var(--text-secondary);
    cursor: pointer;
    transition: color var(--transition-fast), border-color var(--transition-fast), background var(--transition-fast);
  }

  .account-btn:hover,
  .account-btn[aria-expanded='true'] {
    color: var(--accent);
    border-color: var(--accent);
    background: var(--accent-muted);
  }

  .menu-backdrop {
    position: fixed;
    inset: 0;
    z-index: 99;
    background: transparent;
    border: none;
    cursor: default;
  }

  @media (max-width: 480px) {
    .brand-text {
      display: none;
    }
    .user-name {
      display: none;
    }
  }
</style>
