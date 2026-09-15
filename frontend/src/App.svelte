<script lang="ts">
  import Router, { location, push, replace } from 'svelte-spa-router'
  import { wrap } from 'svelte-spa-router/wrap'
  import { onMount } from 'svelte'
  import { auth } from './lib/auth'
  import Navbar from './components/Navbar.svelte'

  import Login from './routes/Login.svelte'
  import Register from './routes/Register.svelte'
  import Home from './routes/Home.svelte'
  import Settings from './routes/Settings.svelte'
  import Help from './routes/Help.svelte'
  import FAQ from './routes/FAQ.svelte'
  import Logs from './routes/Logs.svelte'
  import About from './routes/About.svelte'
  import Account from './routes/Account.svelte'

  const publicPaths = new Set(['/login', '/register'])

  const routes = {
    '/login': Login,
    '/register': Register,
    '/': wrap({
      component: Home,
      conditions: [() => $auth.status === 'authenticated'],
    }),
    '/settings': wrap({
      component: Settings,
      conditions: [() => $auth.status === 'authenticated'],
    }),
    '/help': wrap({
      component: Help,
      conditions: [() => $auth.status === 'authenticated'],
    }),
    '/faq': wrap({
      component: FAQ,
      conditions: [() => $auth.status === 'authenticated'],
    }),
    '/logs': wrap({
      component: Logs,
      conditions: [() => $auth.status === 'authenticated'],
    }),
    '/about': wrap({
      component: About,
      conditions: [() => $auth.status === 'authenticated'],
    }),
    '/account': wrap({
      component: Account,
      conditions: [() => $auth.status === 'authenticated'],
    }),
    // Unknown paths: require auth and land on home (guard redirects if anonymous).
    '*': wrap({
      component: Home,
      conditions: [() => $auth.status === 'authenticated'],
    }),
  }

  let ready = false

  onMount(async () => {
    await auth.refresh()
    ready = true
    enforceAuth($location)
  })

  function enforceAuth(path: string) {
    if (!ready) return
    const isPublic = publicPaths.has(path)
    if ($auth.status === 'authenticated' && isPublic) {
      replace('/')
    } else if ($auth.status !== 'authenticated' && !isPublic) {
      replace('/login')
    }
  }

  $: if (ready) enforceAuth($location)

  function conditionsFailed() {
    push('/login')
  }
</script>

{#if !ready}
  <div class="boot">
    <div class="boot-spinner" aria-label="Loading"></div>
  </div>
{:else}
  {#if $auth.status === 'authenticated'}
    <Navbar />
  {/if}
  <Router {routes} on:conditionsFailed={conditionsFailed} />
{/if}

<style>
  .boot {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--bg-primary);
  }

  .boot-spinner {
    width: 1.5rem;
    height: 1.5rem;
    border: 2px solid var(--border);
    border-top-color: var(--accent);
    border-radius: 50%;
    animation: spin 0.6s linear infinite;
  }
</style>
