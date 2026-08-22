const CELL_STAGGER_MS = 30
const CELL_MAX_DELAY_MS = 240

export function cellDelay(index: number): string {
  return `${Math.min(index * CELL_STAGGER_MS, CELL_MAX_DELAY_MS)}ms`
}
