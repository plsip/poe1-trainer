import type {
  Guide,
  RunSession,
  CurrentState,
  Recommendation,
  Checkpoint,
  RankingEntry,
} from './types'

const BASE = '/api'

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json', ...init?.headers },
    ...init,
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error((body as { error?: string }).error ?? `HTTP ${res.status}`)
  }
  return res.json() as Promise<T>
}

// ─── guides ────────────────────────────────────────────────────────────────

export function listGuides(): Promise<Guide[]> {
  return request('/guides')
}

export function getGuide(slug: string): Promise<Guide> {
  return request(`/guides/${slug}`)
}

// ─── runs ──────────────────────────────────────────────────────────────────

export function createRun(guideId: number, characterName: string): Promise<RunSession> {
  return request('/runs', {
    method: 'POST',
    body: JSON.stringify({ guide_id: guideId, character_name: characterName }),
  })
}

export function getRunState(runId: number): Promise<CurrentState> {
  return request(`/runs/${runId}/state`)
}

export function getRecommendations(runId: number): Promise<Recommendation[]> {
  return request(`/runs/${runId}/recommendations`)
}

export function confirmStep(runId: number, stepId: number): Promise<Checkpoint> {
  return request(`/runs/${runId}/steps/${stepId}/confirm`, { method: 'POST' })
}

export function finishRun(runId: number): Promise<void> {
  return request(`/runs/${runId}/finish`, { method: 'POST' })
}

export function listRuns(guideSlug: string): Promise<RunSession[]> {
  return request(`/guides/${guideSlug}/runs`)
}

export function getRanking(guideSlug: string): Promise<RankingEntry[]> {
  return request(`/guides/${guideSlug}/ranking`)
}
