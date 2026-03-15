import type { GuideStep } from '../api/types'

export function buildActStepNumberMap(steps: GuideStep[]): Map<number, number> {
  const map = new Map<number, number>()
  const counts = new Map<number, number>()

  for (const step of steps) {
    const next = (counts.get(step.act) ?? 0) + 1
    counts.set(step.act, next)
    map.set(step.id, next)
  }

  return map
}

export function formatActStepLabel(step: GuideStep, actStepNumber?: number): string {
  const localStep = actStepNumber ?? step.step_number
  return `A${step.act} · K${localStep}`
}