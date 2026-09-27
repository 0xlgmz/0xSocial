import { createContext } from 'react'
import type { MonetizationConfig } from '../../api/monetization'

export type MonetizationStatus = 'loading' | 'enabled' | 'disabled' | 'unavailable'
export type ConsentStatus = 'checking' | 'ready' | 'unavailable'

export type MonetizationContextValue = {
  config: MonetizationConfig
  status: MonetizationStatus
  consentStatus: ConsentStatus
  canRequestAds: boolean
  openPrivacyChoices: () => boolean
}

export const MonetizationContext = createContext<MonetizationContextValue | null>(null)
