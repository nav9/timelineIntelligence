<script lang="ts">
  import type { PasswordStrength } from '../lib/password'

  export let strength: PasswordStrength = 'weak'
  export let feedback: string[] = []
  /** When true, show advisory feedback even if password is valid. */
  export let showFeedback = true

  const labels: Record<PasswordStrength, string> = {
    weak: 'Weak',
    fair: 'Fair',
    good: 'Good',
    strong: 'Strong',
  }

  $: segments = [
    strength !== 'weak' || strength === 'weak',
    strength === 'fair' || strength === 'good' || strength === 'strong',
    strength === 'good' || strength === 'strong',
    strength === 'strong',
  ]
</script>

{#if strength}
  <div class="strength-bar" aria-hidden="true">
    <div class="strength-segment" class:weak={segments[0] && strength === 'weak'} class:fair={segments[0] && strength === 'fair'} class:good={segments[0] && (strength === 'good' || strength === 'strong')} class:strong={segments[0] && strength === 'strong'}></div>
    <div class="strength-segment" class:fair={segments[1] && strength === 'fair'} class:good={segments[1] && strength === 'good'} class:strong={segments[1] && strength === 'strong'}></div>
    <div class="strength-segment" class:good={segments[2] && strength === 'good'} class:strong={segments[2] && strength === 'strong'}></div>
    <div class="strength-segment" class:strong={segments[3]}></div>
  </div>
  <div class="strength-label {strength}">{labels[strength]}</div>
{/if}

{#if showFeedback && feedback.length > 0}
  <ul class="feedback-list">
    {#each feedback as item}
      <li>{item}</li>
    {/each}
  </ul>
{/if}
