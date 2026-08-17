/**
 * @vitest-environment jsdom
 */
import { describe, expect, it } from 'vitest'
import { ApiError } from './client'
import { formatAuthError } from './errors'

describe('formatAuthError', () => {
  it('maps known auth codes', () => {
    expect(formatAuthError(new ApiError('UNAUTHORIZED', 'Unauthorized', 401), 'fallback')).toBe(
      'Incorrect email or password.',
    )
    expect(formatAuthError(new ApiError('FORBIDDEN', 'Forbidden', 403), 'fallback')).toBe(
      'Invalid invite code.',
    )
    expect(formatAuthError(new ApiError('CONFLICT', 'Conflict', 409), 'fallback')).toBe(
      'An account with this email already exists.',
    )
  })

  it('uses caller fallback for verify/validation instead of generic API copy', () => {
    expect(
      formatAuthError(new ApiError('VALIDATION_ERROR', 'Invalid request', 400), 'Invalid or expired code. Try again.'),
    ).toBe('Invalid or expired code. Try again.')
  })

  it('does not leak raw API messages', () => {
    expect(formatAuthError(new ApiError('INTERNAL', 'panic at line 12', 500), 'Sign in failed. Try again.')).toBe(
      'Sign in failed. Try again.',
    )
  })
})
