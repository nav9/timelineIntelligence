import { vitePreprocess } from '@sveltejs/vite-plugin-svelte'

/** @type {import('@sveltejs/vite-plugin-svelte').SvelteConfig} */
const config = {
  preprocess: vitePreprocess(),
  compilerOptions: {
    // Enable runes if migrating to Svelte 5 later; keep classic for Svelte 4.
  },
}

export default config
