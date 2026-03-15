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

export type StepProgressStatus =
  | 'pending'
  | 'in_progress'
  | 'needs_confirmation'
  | 'completed'
  | 'skipped'

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
  guide_revision: number
  character_name: string
  league?: string
  status: 'active' | 'finished' | 'abandoned'
  started_at: string
  finished_at?: string
  is_active: boolean
  timer_started_at?: string
  paused_at?: string
  total_paused_ms: number
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
  step_timings?: StepTiming[]
}

export interface StepTiming {
  step_id: number
  split_ms: number
  delta_pb_ms?: number
}

export interface Alert {
  kind: 'gem' | 'gear'
  priority: 'high' | 'medium' | 'low'
  slot?: string
  description: string
  step_id?: number
  notes?: string
  // Extended fields (Prompt 07)
  gem_name?: string     // gem name (gem alerts only)
  action_type?: string  // sub-type: vendor | quest_reward | weapon_swap | full_switch | gear_4link | …
  reason?: string       // explanation of why this alert matters now
  source?: 'step' | 'rule' // "step" = step-specific | "rule" = campaign-phase rule
}

export interface AlertsResponse {
  step_id: number
  alerts: Alert[]
}

export interface Split {
  id: number
  run_id: number
  step_id: number
  split_ms: number
  recorded_at: string
}

export interface ManualCheck {
  id: number
  run_id: number
  step_id?: number
  check_type: string
  prompt: string
  is_confirmed: boolean
  response_value?: string
  confirmed_at?: string
  created_at: string
}

export interface StepFilter {
  act: number | null
  status: 'all' | 'completed' | 'pending' | 'current'
  type: 'all' | 'checkpoint' | 'regular'
}

export interface RunEvent {
  id: number
  run_id: number
  event_type: string
  payload: Record<string, string>
  occurred_at: string
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

export interface DetailedRankingEntry {
  rank: number
  run_id: number
  character_name: string
  started_at: string
  total_ms: number
  act_splits: Record<string, number>
  is_pb: boolean
}

export interface IntegrationStatus {
  log_watcher: 'active' | 'waiting_for_file' | 'waiting_for_new_lines' | 'game_not_running' | 'parser_error' | 'disabled'
}

export interface RankingStats {
  count: number
  pb_ms?: number
  median_ms?: number
  last_run_ms?: number
  last_run_id?: number
}

export interface SplitDelta {
  step_id: number
  split_ms: number
  delta_pb_ms?: number
  delta_prev_ms?: number
}

export interface RunDeltasResponse {
  run_id: number
  pb_run_id?: number
  prev_run_id?: number
  splits: SplitDelta[]
}

export interface ReplayLogResult {
  lines_read: number
  events_dispatched: number
  parse_errors: number
  duration_ms: number
}
