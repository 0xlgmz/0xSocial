import { useEffect } from 'react'
import { ensureAdSenseScript } from './adsense-script-state'
import { useMonetization } from './useMonetization'

export function AdSenseScript() {
  const { config, status } = useMonetization()

  useEffect(() => {
    // Google's AdSense tag bootstraps published Privacy & Messaging consent
    // messages and applies the resulting serving mode to individual ad units.
    if (status === 'enabled' && config.clientId) ensureAdSenseScript(config.clientId)
  }, [status, config.clientId])

  return null
}
