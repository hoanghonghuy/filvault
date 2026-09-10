/**
 * @vitest-environment jsdom
 */
import { describe, expect, it } from 'vitest'
import { ApiError } from './client'
import { formatApiError, formatAuthError, formatShareUserError } from './errors'

describe('formatApiError', () => {
  it('maps fetch failures to a reachability message instead of the caller fallback', () => {
    expect(formatApiError(new TypeError('Failed to fetch'), 'Failed to load photos')).toBe(
      "Can't reach the server. Try again in a moment.",
    )
  })

  it('keeps the caller fallback for unknown non-API errors', () => {
    expect(formatApiError(new Error('unexpected json'), 'Failed to load photos')).toBe(
      'Failed to load photos',
    )
  })
})

describe('formatShareUserError', () => {
  it('maps duplicate share conflicts to an actionable message', () => {
    expect(formatShareUserError(new ApiError('CONFLICT', 'Conflict', 409), 'Could not share')).toBe(
      'This item is already shared with that user.',
    )
  })

  it('maps validation errors for share input', () => {
    expect(
      formatShareUserError(new ApiError('VALIDATION_ERROR', 'Invalid request', 400), 'Could not share'),
    ).toBe('Enter a valid email address.')
  })

  it('maps fetch failures to a reachability message', () => {
    expect(formatShareUserError(new TypeError('Failed to fetch'), 'Could not share')).toBe(
      "Can't reach the server. Try again in a moment.",
    )
  })
})

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
