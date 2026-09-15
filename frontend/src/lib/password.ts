/**
 * Client-side password validation for usability feedback.
 * Backend validation is authoritative — this mirrors the policy for UX only.
 */

export type PasswordStrength = 'weak' | 'fair' | 'good' | 'strong'

export interface PasswordValidation {
  valid: boolean
  strength: PasswordStrength
  feedback: string[]
}

const COMMON = new Set([
  'password', 'password1', 'password123',
  '123456', '1234567', '12345678', '123456789',
  'qwerty', 'qwerty123', 'abc123', 'letmein',
  'welcome', 'monkey', 'dragon', 'master',
  'sunshine', 'princess', 'shadow', 'superman',
  'iloveyou', 'trustno1', 'admin', 'login',
  'passw0rd', 'p@ssword', 'p@ssw0rd', 'secret',
])

function isSequential(s: string): boolean {
  if (s.length < 6) return false
  let sequential = 0
  for (let i = 1; i < s.length; i++) {
    const diff = s.charCodeAt(i) - s.charCodeAt(i - 1)
    if (diff === 1 || diff === -1) sequential++
  }
  return sequential >= s.length - 2
}

function assessStrength(password: string): PasswordStrength {
  const length = password.length
  let variety = 0
  if (/[a-z]/.test(password)) variety++
  if (/[A-Z]/.test(password)) variety++
  if (/[0-9]/.test(password)) variety++
  if (/[^a-zA-Z0-9]/.test(password)) variety++

  if (length >= 20 && variety >= 3) return 'strong'
  if (length >= 14 && variety >= 2) return 'good'
  if (length >= 10) return 'fair'
  return 'weak'
}

/**
 * Validate a password for client-side feedback.
 * Does not replace server-side checks.
 */
export function validatePassword(
  password: string,
  name = '',
  email = '',
): PasswordValidation {
  const feedback: string[] = []

  if (password.length < 10) {
    feedback.push('Password must be at least 10 characters long. Consider using a passphrase.')
  }

  if (COMMON.has(password.toLowerCase())) {
    feedback.push('This password is too common. Please choose something more unique.')
  }

  if (/^(.)\1{5,}$/.test(password)) {
    feedback.push('Avoid passwords made of repeated characters.')
  }

  if (isSequential(password)) {
    feedback.push("Avoid obvious sequences like '12345678' or 'abcdef'.")
  }

  const lower = password.toLowerCase()
  if (name && name.length >= 4 && lower.includes(name.toLowerCase())) {
    feedback.push('Avoid using your name in your password.')
  }
  if (email) {
    const local = email.split('@')[0] || ''
    if (local.length >= 4 && lower.includes(local.toLowerCase())) {
      feedback.push('Avoid using your email address in your password.')
    }
  }

  const strength = assessStrength(password)
  const valid = feedback.length === 0

  if (valid && strength === 'fair') {
    feedback.push('Password is acceptable but could be stronger. Consider a longer passphrase.')
  }

  return { valid, strength, feedback }
}

export function isValidEmail(email: string): boolean {
  const parts = email.split('@')
  if (parts.length !== 2) return false
  const [local, domain] = parts
  if (!local || !domain) return false
  return domain.includes('.')
}
