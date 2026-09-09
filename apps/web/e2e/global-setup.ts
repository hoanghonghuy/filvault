const apiURL = process.env.E2E_API_URL ?? 'http://127.0.0.1:8080'

async function waitForApi(): Promise<void> {
  const deadline = Date.now() + 60_000
  let lastError = 'API did not become ready'

  while (Date.now() < deadline) {
    try {
      const response = await fetch(`${apiURL}/healthz`)
      if (response.ok) {
        return
      }
      lastError = `healthz returned ${response.status}`
    } catch (error) {
      lastError = error instanceof Error ? error.message : String(error)
    }
    await new Promise((resolve) => setTimeout(resolve, 1_000))
  }

  throw new Error(`E2E API readiness check failed: ${lastError}`)
}

export default async function globalSetup(): Promise<void> {
  await waitForApi()
}
