import { useEffect, useRef, useState } from 'react'
import type { FeedAdPlacement } from '../../api/monetization'
import { useAdSenseScriptStatus } from '../../features/monetization/adsense-script-state'

declare global {
  interface Window {
    adsbygoogle?: Record<string, unknown>[]
  }
}

export function FeedAdCard({ clientId, placement }: { clientId: string; placement: FeedAdPlacement }) {
  const scriptStatus = useAdSenseScriptStatus()
  const requested = useRef(false)
  const unit = useRef<HTMLModElement>(null)
  const [unfilled, setUnfilled] = useState(false)

  useEffect(() => {
    if (scriptStatus !== 'ready' || requested.current) return
    requested.current = true

    try {
      window.adsbygoogle ??= []
      window.adsbygoogle.push({})
    } catch {
      window.queueMicrotask(() => setUnfilled(true))
    }
  }, [scriptStatus])

  useEffect(() => {
    const element = unit.current
    if (!element) return
    const observer = new MutationObserver(() => {
      if (element.dataset.adStatus === 'unfilled') setUnfilled(true)
    })
    observer.observe(element, { attributes: true, attributeFilter: ['data-ad-status'] })
    return () => observer.disconnect()
  }, [])

  if (scriptStatus === 'failed' || unfilled) return null

  return (
    <aside className="card min-h-44 border border-base-300 bg-base-100 shadow-sm sm:min-h-52" aria-label="Advertisement">
      <div className="card-body gap-2 p-4 sm:p-5">
        <span className="text-[.65rem] uppercase tracking-wide text-base-content/40">Advertisement</span>
        <ins
          ref={unit}
          className="adsbygoogle min-h-32 w-full"
          style={{ display: 'block' }}
          data-ad-client={clientId}
          data-ad-slot={placement.slotId}
          data-ad-format={placement.format}
          data-ad-layout-key={placement.layoutKey}
        />
      </div>
    </aside>
  )
}
