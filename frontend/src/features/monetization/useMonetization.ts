import { useContext } from 'react'
import { MonetizationContext } from './monetization-context'

export function useMonetization() {
  const context = useContext(MonetizationContext)
  if (!context) throw new Error('useMonetization must be used inside MonetizationProvider.')
  return context
}
