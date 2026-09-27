import { AuthProvider } from '../features/auth/AuthProvider'
import { AdSenseScript } from '../features/monetization/AdSenseScript'
import { MonetizationProvider } from '../features/monetization/MonetizationProvider'
import '../styles/brand.css'
import { AppRouter } from './AppRouter'

export function App() {
  return <MonetizationProvider><AdSenseScript/><AuthProvider><AppRouter/></AuthProvider></MonetizationProvider>
}
