/** Motion tokens aligned with DESIGN.md (150–250ms) and Material 3 Standard (web). */
export const MOTION_PRESS_MS = 150
export const MOTION_ENTER_MS = 250
export const MOTION_EXIT_MS = 200

export const EASE_STANDARD = 'cubic-bezier(0.2, 0, 0, 1)'
export const EASE_ENTER = 'cubic-bezier(0, 0, 0, 1)'
export const EASE_EXIT = 'cubic-bezier(0.3, 0, 1, 1)'

const CELL_STAGGER_MS = 30
const CELL_MAX_DELAY_MS = 240

/** Staggered delay for grid/list cell entrance animations (Photos, Album views). */
export function cellDelay(index: number): string {
  return `${Math.min(index * CELL_STAGGER_MS, CELL_MAX_DELAY_MS)}ms`
}
