import { useEffect } from 'react'
import { ensureAdSenseScript } from './adsense-script-state'
import { useMonetization } from './useMonetization'

export function AdSenseScript() {
  const { config, canRequestAds } = useMonetization()

  useEffect(() => {
    if (canRequestAds && config.clientId) ensureAdSenseScript(config.clientId)
  }, [canRequestAds, config.clientId])

  return null
}
