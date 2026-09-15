<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from 'svelte'

  const dispatch = createEventDispatcher<{
    navigate: string
    logout: void
    close: void
  }>()

  const items: { label: string; path?: string; action?: 'logout'; danger?: boolean }[] = [
    { label: 'Settings', path: '/settings' },
    { label: 'Help', path: '/help' },
    { label: 'FAQ', path: '/faq' },
    { label: 'Logs', path: '/logs' },
    { label: 'About', path: '/about' },
    { label: 'Unregister / Deactivate Account', path: '/account', danger: true },
    { label: 'Logout', action: 'logout' },
  ]

  function onKey(e: KeyboardEvent) {
    if (e.key === 'Escape') dispatch('close')
  }

  onMount(() => {
    window.addEventListener('keydown', onKey)
  })

  onDestroy(() => {
    window.removeEventListener('keydown', onKey)
  })

  function activate(item: (typeof items)[0]) {
    if (item.action === 'logout') {
      dispatch('logout')
    } else if (item.path) {
      dispatch('navigate', item.path)
    }
  }
</script>

<div class="menu" role="menu" aria-label="Account menu">
  {#each items as item, i}
    {#if i === 5}
      <div class="menu-sep" role="separator"></div>
    {/if}
    <button
      type="button"
      class="menu-item"
      class:danger={item.danger}
      role="menuitem"
      on:click={() => activate(item)}
    >
      {item.label}
    </button>
  {/each}
</div>

<style>
  .menu {
    position: absolute;
    top: calc(100% + 4px);
    right: 0;
    min-width: 220px;
    background: var(--bg-elevated);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    box-shadow: var(--shadow);
    padding: var(--sp-1) 0;
    z-index: 110;
  }

  .menu-item {
    display: block;
    width: 100%;
    text-align: left;
    padding: var(--sp-2) var(--sp-4);
    background: none;
    border: none;
    color: var(--text-primary);
    font-family: var(--font-sans);
    font-size: var(--text-sm);
    cursor: pointer;
    transition: background var(--transition-fast);
  }

  .menu-item:hover,
  .menu-item:focus {
    background: var(--bg-tertiary);
    outline: none;
  }

  .menu-item.danger {
    color: var(--error);
  }

  .menu-sep {
    height: 1px;
    background: var(--border);
    margin: var(--sp-1) 0;
  }
</style>
