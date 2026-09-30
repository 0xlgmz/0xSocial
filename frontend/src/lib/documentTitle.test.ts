import { describe, expect, it } from 'vitest'
import { getDocumentTitle } from './documentTitle'

describe('getDocumentTitle', () => {
  it('uses only the site name on the homepage', () => {
    expect(getDocumentTitle({ kind: 'static', path: '/' })).toBe('0xSocial')
  })

  it('identifies static pages while retaining the site name', () => {
    expect(getDocumentTitle({ kind: 'static', path: '/explore' })).toBe('Explore | 0xSocial')
    expect(getDocumentTitle({ kind: 'static', path: '/profile/edit' })).toBe('Edit Profile | 0xSocial')
  })

  it('identifies public profiles by handle', () => {
    expect(getDocumentTitle({ kind: 'profile', handle: 'alice' })).toBe('@alice | 0xSocial')
  })

  it('labels unknown endpoints', () => {
    expect(getDocumentTitle({ kind: 'not-found' })).toBe('Page Not Found | 0xSocial')
  })
})
