import { describe, it, expect } from 'vitest'
import { validatePassword, isValidEmail } from '../src/lib/password'

describe('isValidEmail', () => {
  it('accepts a normal email', () => {
    expect(isValidEmail('user@example.com')).toBe(true)
  })

  it('rejects missing @', () => {
    expect(isValidEmail('not-an-email')).toBe(false)
  })

  it('rejects domain without dot', () => {
    expect(isValidEmail('user@localhost')).toBe(false)
  })
})

describe('validatePassword', () => {
  it('rejects short passwords', () => {
    const r = validatePassword('short')
    expect(r.valid).toBe(false)
    expect(r.feedback.some((f) => f.includes('10 characters'))).toBe(true)
  })

  it('rejects common passwords', () => {
    const r = validatePassword('password123')
    expect(r.valid).toBe(false)
  })

  it('rejects passwords containing the user name', () => {
    const r = validatePassword('alice-secret-ok', 'Alice', 'other@example.com')
    expect(r.valid).toBe(false)
  })

  it('accepts a long passphrase', () => {
    const r = validatePassword('correct-horse-battery-staple')
    expect(r.valid).toBe(true)
    expect(['fair', 'good', 'strong']).toContain(r.strength)
  })

  it('rejects obvious sequences', () => {
    const r = validatePassword('abcdefghij')
    expect(r.valid).toBe(false)
  })
})
