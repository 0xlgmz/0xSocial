import { StrictMode } from 'react'
import { cleanup, render, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { disabledMonetizationConfig, normalizeMonetizationConfig } from '../../api/monetization'
import type { FeedAdPlacement } from '../../api/monetization'
import { FeedAdCard } from '../../components/ads/FeedAdCard'
import { AdSenseScript } from './AdSenseScript'
import { ensureAdSenseScript } from './adsense-script-state'
import { interleaveFeedAds } from './interleaveFeedAds'
import { MonetizationContext } from './monetization-context'

const placement: FeedAdPlacement = {
  enabled: true,
  slotId: '1234567890',
  format: 'fluid',
  layoutKey: '-fb+5w+4e-db+86',
  firstAfterPosts: 4,
  repeatEveryPosts: 8,
  maximumAds: 3,
}

afterEach(() => cleanup())

describe('monetization configuration', () => {
  it('disables advertising globally when the backend disables it', () => {
    expect(normalizeMonetizationConfig({ enabled: false })).toEqual(disabledMonetizationConfig())
  })

  it('keeps a disabled feed placement disabled', () => {
    const config = normalizeMonetizationConfig({
      enabled: true,
      clientId: 'ca-pub-1234567890123456',
      placements: { feed: { enabled: false } },
    })
    expect(config.enabled).toBe(true)
    expect(config.placements.feed.enabled).toBe(false)
  })

  it.each([
    null,
    {},
    { enabled: true },
    { enabled: true, clientId: 'wrong', placements: { feed: placement } },
    { enabled: true, clientId: 'ca-pub-1234567890123456', placements: { feed: { ...placement, maximumAds: -1 } } },
  ])('turns malformed configuration into a safe disabled configuration', value => {
    expect(normalizeMonetizationConfig(value).enabled).toBe(false)
  })
})

describe('feed interleaving', () => {
  const posts = Array.from({ length: 24 }, (_, index) => `post-${index + 1}`)
  const interleave = (items = posts, config = placement) => interleaveFeedAds(
    items,
    config,
    post => post,
    index => `ad-${index + 1}`,
  )

  it('adds no ad to an empty feed', () => expect(interleave([])).toEqual([]))

  it('adds no ad when the placement is disabled', () => {
    expect(interleave(posts, { ...placement, enabled: false })).toEqual(posts)
  })

  it('places the first ad after the configured post count', () => {
    expect(interleave().slice(0, 6)).toEqual(['post-1', 'post-2', 'post-3', 'post-4', 'ad-1', 'post-5'])
  })

  it('uses the repeat interval and enforces the maximum count', () => {
    const result = interleave()
    expect(result.filter(item => typeof item === 'string' && item.startsWith('ad-'))).toEqual(['ad-1', 'ad-2', 'ad-3'])
    expect(result.indexOf('ad-2')).toBe(13)
    expect(result.indexOf('ad-3')).toBe(22)
  })

  it('adds no ad before enough posts have rendered', () => {
    expect(interleave(posts.slice(0, 3))).toEqual(posts.slice(0, 3))
  })
})

describe('AdSense lifecycle', () => {
  it('loads the AdSense tag while consent is still being resolved', async () => {
    const config = normalizeMonetizationConfig({
      enabled: true,
      clientId: 'ca-pub-1234567890123456',
      placements: { feed: placement },
    })

    render(
      <MonetizationContext.Provider value={{
        config,
        status: 'enabled',
        consentStatus: 'checking',
        canRequestAds: false,
        openPrivacyChoices: () => false,
      }}>
        <AdSenseScript/>
      </MonetizationContext.Provider>,
    )

    await waitFor(() => {
      const script = document.querySelector<HTMLScriptElement>('script[data-0xsocial-adsense]')
      expect(script?.src).toContain('client=ca-pub-1234567890123456')
    })
  })

  it('adds the script and requests a mounted unit only once under StrictMode and rerenders', async () => {
    ensureAdSenseScript('ca-pub-1234567890123456')
    ensureAdSenseScript('ca-pub-1234567890123456')
    const scripts = document.querySelectorAll('script[data-0xsocial-adsense]')
    expect(scripts).toHaveLength(1)
    scripts[0].dispatchEvent(new Event('load'))

    window.adsbygoogle = []
    const view = render(<StrictMode><FeedAdCard clientId="ca-pub-1234567890123456" placement={placement}/></StrictMode>)
    await waitFor(() => expect(window.adsbygoogle).toHaveLength(1))
    view.rerender(<StrictMode><FeedAdCard clientId="ca-pub-1234567890123456" placement={placement}/></StrictMode>)
    expect(window.adsbygoogle).toHaveLength(1)
  })

  it('contains an ad request failure without throwing', async () => {
    window.adsbygoogle = { push: () => { throw new Error('blocked') } } as unknown as Record<string, unknown>[]
    const view = render(<FeedAdCard clientId="ca-pub-1234567890123456" placement={placement}/>)
    await waitFor(() => expect(view.queryByLabelText('Advertisement')).not.toBeInTheDocument())
  })
})
