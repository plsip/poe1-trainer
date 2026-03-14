import type {
  Guide,
  RunSession,
  CurrentState,
  Recommendation,
  Checkpoint,
  DetailedRankingEntry,
  RankingStats,
  AlertsResponse,
  Split,
  ManualCheck,
  RunEvent,
  RunDeltasResponse,
  IntegrationStatus,
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

  if (res.status === 204) {
    return undefined as T
  }

  const contentLength = res.headers.get('Content-Length')
  if (contentLength === '0') {
    return undefined as T
  }

  const contentType = res.headers.get('Content-Type') ?? ''
  if (!contentType.includes('application/json')) {
    return undefined as T
  }

  const text = await res.text()
  if (!text.trim()) {
    return undefined as T
  }

  return JSON.parse(text) as T
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

export function getRunGuide(runId: number): Promise<Guide> {
  return request(`/runs/${runId}/guide`)
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

export function listActiveRuns(): Promise<RunSession[]> {
  return request('/runs/active')
}

export function getRanking(guideSlug: string): Promise<DetailedRankingEntry[]> {
  return request(`/guides/${guideSlug}/ranking`)
}

export function getRankingStats(guideSlug: string): Promise<RankingStats> {
  return request(`/guides/${guideSlug}/ranking/stats`)
}

// ─── extended run actions ──────────────────────────────────────────────────

export function skipStep(runId: number, stepId: number): Promise<void> {
  return request(`/runs/${runId}/steps/${stepId}/skip`, { method: 'POST' })
}

export function undoStep(runId: number, stepId: number): Promise<void> {
  return request(`/runs/${runId}/steps/${stepId}/undo`, { method: 'POST' })
}

export function abandonRun(runId: number): Promise<void> {
  return request(`/runs/${runId}/abandon`, { method: 'POST' })
}

// ─── alerts ────────────────────────────────────────────────────────────────

export function getAlerts(runId: number): Promise<AlertsResponse> {
  return request(`/runs/${runId}/alerts`)
}

// ─── splits ────────────────────────────────────────────────────────────────

export function getSplits(runId: number): Promise<Split[]> {
  return request(`/runs/${runId}/splits`)
}

export function getSplitDeltas(runId: number): Promise<RunDeltasResponse> {
  return request(`/runs/${runId}/split-deltas`)
}

// ─── pause / resume ────────────────────────────────────────────────────────

export function pauseRun(runId: number): Promise<void> {
  return request(`/runs/${runId}/pause`, { method: 'POST' })
}

export function resumeRun(runId: number): Promise<void> {
  return request(`/runs/${runId}/resume`, { method: 'POST' })
}

// ─── manual checks ─────────────────────────────────────────────────────────

export function getChecks(runId: number): Promise<ManualCheck[]> {
  return request(`/runs/${runId}/checks`)
}

export function answerCheck(runId: number, checkId: number, value: string): Promise<void> {
  return request(`/runs/${runId}/checks/${checkId}/answer`, {
    method: 'POST',
    body: JSON.stringify({ response_value: value }),
  })
}

// ─── integration status ────────────────────────────────────────────────────

export function getIntegrationStatus(): Promise<IntegrationStatus> {
  return request('/integration/status')
}

// ─── events ────────────────────────────────────────────────────────────────

export function listEvents(runId: number): Promise<RunEvent[]> {
  return request(`/runs/${runId}/events`)
}

// ─── SSE streams ───────────────────────────────────────────────────────────

/**
 * Opens an SSE stream that pushes raw Client.txt lines as they are read by the
 * backend watcher.  Events: "status" (initial watcher status), "log_line".
 */
export function subscribeToLogTail(): EventSource {
  return new EventSource(`${BASE}/integration/logtail/stream`)
}

/**
 * Opens an SSE stream that pushes run state updates for the given run.
 * Events: "state" (CurrentState JSON).  A keep-alive comment is sent every 30 s.
 */
export function subscribeToRunStream(runId: number): EventSource {
  return new EventSource(`${BASE}/runs/${runId}/stream`)
}
