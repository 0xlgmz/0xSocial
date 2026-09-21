export type HealthResponse = {
  status: string
}

export async function getHealth(
  signal?: AbortSignal,
): Promise<HealthResponse> {
  const response = await fetch('/api/health', { signal })

  if (!response.ok) {
    throw new Error(
      `Health request failed: ${response.status} ${response.statusText}`,
    )
  }

  return response.json() as Promise<HealthResponse>
}