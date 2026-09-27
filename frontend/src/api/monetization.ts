import { request } from './client'

export type FeedAdPlacement = {
  enabled: boolean
  slotId: string
  format: string
  layoutKey?: string
  firstAfterPosts: number
  repeatEveryPosts: number
  maximumAds: number
}

export type MonetizationConfig = {
  enabled: boolean
  clientId: string
  placements: { feed: FeedAdPlacement }
}

export type MonetizationConfigResult = {
  config: MonetizationConfig
  available: boolean
}

const disabledFeed: FeedAdPlacement = {
  enabled: false,
  slotId: '',
  format: 'fluid',
  firstAfterPosts: 0,
  repeatEveryPosts: 0,
  maximumAds: 0,
}

export function disabledMonetizationConfig(): MonetizationConfig {
  return { enabled: false, clientId: '', placements: { feed: { ...disabledFeed } } }
}

function record(value: unknown): Record<string, unknown> | null {
  return value !== null && typeof value === 'object' && !Array.isArray(value)
    ? value as Record<string, unknown>
    : null
}

function positiveInteger(value: unknown): value is number {
  return typeof value === 'number' && Number.isSafeInteger(value) && value > 0
}

export function normalizeMonetizationConfig(value: unknown): MonetizationConfig {
  const root = record(value)
  if (!root || root.enabled !== true) return disabledMonetizationConfig()

  const clientId = root.clientId
  const placements = record(root.placements)
  const feed = record(placements?.feed)
  if (typeof clientId !== 'string' || !/^ca-pub-\d{16}$/.test(clientId) || !feed) {
    return disabledMonetizationConfig()
  }

  if (feed.enabled !== true) {
    return { enabled: true, clientId, placements: { feed: { ...disabledFeed } } }
  }

  const { slotId, format, layoutKey, firstAfterPosts, repeatEveryPosts, maximumAds } = feed
  if (
    typeof slotId !== 'string' || !/^\d+$/.test(slotId)
    || typeof format !== 'string' || format.length === 0
    || (layoutKey !== undefined && typeof layoutKey !== 'string')
    || !positiveInteger(firstAfterPosts)
    || !positiveInteger(repeatEveryPosts)
    || !positiveInteger(maximumAds)
  ) return disabledMonetizationConfig()

  return {
    enabled: true,
    clientId,
    placements: {
      feed: { enabled: true, slotId, format, layoutKey, firstAfterPosts, repeatEveryPosts, maximumAds },
    },
  }
}

export const monetizationApi = {
  async getConfig(signal?: AbortSignal): Promise<MonetizationConfig> {
    try {
      const response = await request<unknown>('/api/monetization/config', { method: 'GET', signal })
      return normalizeMonetizationConfig(response)
    } catch (error) {
      if (error instanceof DOMException && error.name === 'AbortError') throw error
      return disabledMonetizationConfig()
    }
  },

  async getConfigResult(signal?: AbortSignal): Promise<MonetizationConfigResult> {
    try {
      const response = await request<unknown>('/api/monetization/config', { method: 'GET', signal })
      return { config: normalizeMonetizationConfig(response), available: true }
    } catch (error) {
      if (error instanceof DOMException && error.name === 'AbortError') throw error
      return { config: disabledMonetizationConfig(), available: false }
    }
  },
}
