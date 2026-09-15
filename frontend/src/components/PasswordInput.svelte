<script lang="ts">
  export let value = ''
  export let id = 'password'
  export let name = 'password'
  export let placeholder = 'Password'
  export let disabled = false
  export let autocomplete: string = 'current-password'
  export let required = true
  export let error = false

  let visible = false

  function toggle() {
    visible = !visible
  }
</script>

<div class="password-field">
  {#if visible}
    <input
      {id}
      {name}
      type="text"
      class="form-input"
      class:error
      bind:value
      {placeholder}
      {disabled}
      {autocomplete}
      {required}
      aria-label="Password"
    />
  {:else}
    <input
      {id}
      {name}
      type="password"
      class="form-input"
      class:error
      bind:value
      {placeholder}
      {disabled}
      {autocomplete}
      {required}
      aria-label="Password"
    />
  {/if}
  <button
    type="button"
    class="password-toggle"
    on:click={toggle}
    tabindex="0"
    aria-label={visible ? 'Hide password' : 'Show password'}
    title={visible ? 'Hide password' : 'Show password'}
  >
    {#if visible}
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94"/>
        <path d="M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19"/>
        <path d="M14.12 14.12a3 3 0 1 1-4.24-4.24"/>
        <line x1="1" y1="1" x2="23" y2="23"/>
      </svg>
    {:else}
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
        <circle cx="12" cy="12" r="3"/>
      </svg>
    {/if}
  </button>
</div>
