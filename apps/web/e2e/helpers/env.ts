export const e2eEmail = process.env.E2E_EMAIL ?? process.env.FILVAULT_SEED_EMAIL ?? 'dev@filvault.com'

export const e2ePassword =
  process.env.E2E_PASSWORD ?? process.env.FILVAULT_SEED_PASSWORD ?? 'replace-with-local-seed-password'

export const e2eRunId = process.env.E2E_RUN_ID ?? `${Date.now()}`

export const e2ePrefix = `e2e-smoke-${e2eRunId}`

export const e2eFolderName = `${e2ePrefix}-folder`

export const e2eTextFileName = `${e2ePrefix}.txt`

export const e2eImageFileName = `${e2ePrefix}.png`
