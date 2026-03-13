// Types mirroring backend JSON responses.

export interface Guide {
  id: number
  slug: string
  title: string
  build_name: string
  version: string
  created_at: string
  steps?: GuideStep[]
}

export interface GemRequirement {
  id: number
  step_id: number
  gem_name: string
  color: string
  note: string
}

export interface GuideStep {
  id: number
  guide_id: number
  step_number: number
  act: number
  title: string
  description: string
  area: string
  is_checkpoint: boolean
  requires_manual: boolean
  sort_order: number
  gem_requirements?: GemRequirement[]
}

export interface RunSession {
  id: number
  guide_id: number
  character_name: string
  started_at: string
  finished_at?: string
  is_active: boolean
}

export interface Checkpoint {
  id: number
  run_id: number
  step_id: number
  confirmed_at: string
  confirmed_by: string
}

export interface CurrentState {
  run: RunSession
  current_step_id: number
  confirmed_step_ids: number[]
  elapsed_ms: number
}

export interface Recommendation {
  id: string
  text: string
  reason: string
  priority: 'high' | 'medium' | 'low'
  step_id?: number
}

export interface RankingEntry {
  run_id: number
  character_name: string
  started_at: string
  total_ms: number
}
