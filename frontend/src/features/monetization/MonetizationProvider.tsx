import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import type { ReactNode } from 'react'
import { disabledMonetizationConfig, monetizationApi } from '../../api/monetization'
import type { MonetizationConfigResult } from '../../api/monetization'
import { MonetizationContext } from './monetization-context'
import type { ConsentStatus, MonetizationStatus } from './monetization-context'

type TcData = { eventStatus?: string; gdprApplies?: boolean; tcString?: string }
type TcfApi = (command: string, version: number, callback: (data: TcData, success: boolean) => void) => void
type GoogleFcCallback = Record<string, () => void> | (() => void)
type GoogleFc = {
  callbackQueue?: GoogleFcCallback[]
  showRevocationMessage?: () => void
}

declare global {
  interface Window {
    __tcfapi?: TcfApi
    googlefc?: GoogleFc
  }
}

let configRequest: Promise<MonetizationConfigResult> | undefined

function loadConfigOnce() {
  configRequest ??= monetizationApi.getConfigResult()
  return configRequest
}

export function MonetizationProvider({ children }: { children: ReactNode }) {
  const [result, setResult] = useState<MonetizationConfigResult>({
    config: disabledMonetizationConfig(),
    available: true,
  })
  const [status, setStatus] = useState<MonetizationStatus>('loading')
  const [consentStatus, setConsentStatus] = useState<ConsentStatus>('checking')
  const consentResolved = useRef(false)

  useEffect(() => {
    let active = true
    void loadConfigOnce().then(next => {
      if (!active) return
      setResult(next)
      setStatus(!next.available ? 'unavailable' : next.config.enabled ? 'enabled' : 'disabled')
    })
    return () => { active = false }
  }, [])

  useEffect(() => {
    if (status !== 'enabled') return

    const resolveConsent = () => {
      const tcfApi = window.__tcfapi
      if (!tcfApi) return
      tcfApi('addEventListener', 0, (data, success) => {
        if (!success || !data) return
        if (data.gdprApplies === false) {
          consentResolved.current = true
          setConsentStatus('ready')
          return
        }
        const decisionReady = data.eventStatus === 'tcloaded' || data.eventStatus === 'useractioncomplete'
        if (data.gdprApplies === true && decisionReady && typeof data.tcString === 'string' && data.tcString.length > 0) {
          consentResolved.current = true
          setConsentStatus('ready')
        }
      })
    }

    window.googlefc ??= {}
    window.googlefc.callbackQueue ??= []
    window.googlefc.callbackQueue.push({ CONSENT_API_READY: resolveConsent })
    resolveConsent()

    const timeout = window.setTimeout(() => {
      if (!consentResolved.current) setConsentStatus('unavailable')
    }, 8000)
    return () => window.clearTimeout(timeout)
  }, [status])

  const openPrivacyChoices = useCallback(() => {
    const googlefc = window.googlefc
    if (!googlefc?.callbackQueue || !googlefc.showRevocationMessage) return false
    googlefc.callbackQueue.push(googlefc.showRevocationMessage)
    return true
  }, [])

  const value = useMemo(() => ({
    config: result.config,
    status,
    consentStatus,
    canRequestAds: status === 'enabled' && consentStatus === 'ready',
    openPrivacyChoices,
  }), [result.config, status, consentStatus, openPrivacyChoices])

  return <MonetizationContext.Provider value={value}>{children}</MonetizationContext.Provider>
}
