import { useSyncExternalStore } from 'react'

export type AdSenseScriptStatus = 'idle' | 'loading' | 'ready' | 'failed'

let status: AdSenseScriptStatus = 'idle'
const listeners = new Set<() => void>()

function setStatus(next: AdSenseScriptStatus) {
  status = next
  listeners.forEach(listener => listener())
}

export function ensureAdSenseScript(clientId: string) {
  if (status === 'loading' || status === 'ready' || status === 'failed') return

  const existing = document.querySelector<HTMLScriptElement>(
    'script[data-0xsocial-adsense], script[src^="https://pagead2.googlesyndication.com/pagead/js/adsbygoogle.js"]',
  )
  if (existing) {
    if (existing.dataset.loaded === 'true' || window.adsbygoogle) {
      setStatus('ready')
    } else {
      setStatus('loading')
      existing.addEventListener('load', () => setStatus('ready'), { once: true })
      existing.addEventListener('error', () => setStatus('failed'), { once: true })
    }
    return
  }

  setStatus('loading')
  const script = document.createElement('script')
  script.async = true
  script.crossOrigin = 'anonymous'
  script.src = `https://pagead2.googlesyndication.com/pagead/js/adsbygoogle.js?client=${encodeURIComponent(clientId)}`
  script.dataset['0xsocialAdsense'] = 'true'
  script.addEventListener('load', () => {
    script.dataset.loaded = 'true'
    setStatus('ready')
  }, { once: true })
  script.addEventListener('error', () => setStatus('failed'), { once: true })
  document.head.append(script)
}

function subscribe(listener: () => void) {
  listeners.add(listener)
  return () => listeners.delete(listener)
}

export function useAdSenseScriptStatus() {
  return useSyncExternalStore(subscribe, () => status, () => 'idle')
}
