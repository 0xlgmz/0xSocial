import { useEffect, useState } from 'react'
import { getHealth, type HealthResponse } from './api/health'

type HealthState =
  | { status: 'loading' }
  | { status: 'success'; data: HealthResponse }
  | { status: 'error'; message: string }

function App() {
  const [health, setHealth] = useState<HealthState>({
    status: 'loading',
  })

  useEffect(() => {
    const controller = new AbortController()

    async function loadHealth() {
      try {
        const data = await getHealth(controller.signal)

        setHealth({
          status: 'success',
          data,
        })
      } catch (error) {
        if (controller.signal.aborted) {
          return
        }

        setHealth({
          status: 'error',
          message:
            error instanceof Error
              ? error.message
              : 'An unknown error occurred',
        })
      }
    }

    void loadHealth()

    return () => {
      controller.abort()
    }
  }, [])

  if (health.status === 'loading') {
    return <p>Checking API…</p>
  }

  if (health.status === 'error') {
    return <p role="alert">API error: {health.message}</p>
  }

  return <p>API status: {health.data.status}</p>
}

export default App