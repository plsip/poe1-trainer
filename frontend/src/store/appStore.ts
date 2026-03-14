import { create } from 'zustand'
import type {
  Guide,
  CurrentState,
  Recommendation,
  RunSession,
  Alert,
  Split,
  ManualCheck,
  StepFilter,
} from '../api/types'
import * as api from '../api/client'

interface AppState {
  // guide list
  guides: Guide[]
  guidesLoading: boolean

  // active guide
  activeGuide: Guide | null

  // active run
  activeRun: RunSession | null
  runState: CurrentState | null
  recommendations: Recommendation[]
  stateLoading: boolean

  // alerts for current step
  alerts: Alert[]
  alertsLoading: boolean

  // splits for active run
  splits: Split[]

  // manual checks
  checks: ManualCheck[]

  // step list filters
  stepFilter: StepFilter

  // errors
  error: string | null

  // actions
  loadGuides: () => Promise<void>
  loadGuide: (slug: string) => Promise<void>
  startRun: (guideId: number, characterName: string) => Promise<void>
  loadRunState: (runId: number) => Promise<void>
  loadAlerts: (runId: number) => Promise<void>
  loadSplits: (runId: number) => Promise<void>
  loadChecks: (runId: number) => Promise<void>
  confirmStep: (runId: number, stepId: number) => Promise<void>
  skipStep: (runId: number, stepId: number) => Promise<void>
  undoStep: (runId: number, stepId: number) => Promise<void>
  finishRun: (runId: number) => Promise<void>
  abandonRun: (runId: number) => Promise<void>
  answerCheck: (runId: number, checkId: number, value: string) => Promise<void>
  setStepFilter: (patch: Partial<StepFilter>) => void
  clearError: () => void
}

export const useAppStore = create<AppState>((set, get) => ({
  guides: [],
  guidesLoading: false,
  activeGuide: null,
  activeRun: null,
  runState: null,
  recommendations: [],
  stateLoading: false,
  alerts: [],
  alertsLoading: false,
  splits: [],
  checks: [],
  stepFilter: { act: null, status: 'all', type: 'all' },
  error: null,

  loadGuides: async () => {
    set({ guidesLoading: true, error: null })
    try {
      const guides = await api.listGuides()
      set({ guides: guides ?? [], guidesLoading: false })
    } catch (e) {
      set({ guidesLoading: false, error: String(e) })
    }
  },

  loadGuide: async (slug) => {
    set({ error: null })
    try {
      const guide = await api.getGuide(slug)
      set({ activeGuide: guide })
    } catch (e) {
      set({ error: String(e) })
    }
  },

  startRun: async (guideId, characterName) => {
    set({ error: null })
    try {
      const run = await api.createRun(guideId, characterName)
      set({ activeRun: run })
      await get().loadRunState(run.id)
    } catch (e) {
      set({ error: String(e) })
    }
  },

  loadRunState: async (runId) => {
    set({ stateLoading: true, error: null })
    try {
      const [state, recs, guide] = await Promise.all([
        api.getRunState(runId),
        api.getRecommendations(runId),
        api.getRunGuide(runId),
      ])
      set({ runState: state, recommendations: recs ?? [], activeGuide: guide, stateLoading: false })
    } catch (e) {
      set({ stateLoading: false, error: String(e) })
    }
  },

  loadAlerts: async (runId) => {
    set({ alertsLoading: true })
    try {
      const resp = await api.getAlerts(runId)
      set({ alerts: resp.alerts ?? [], alertsLoading: false })
    } catch {
      set({ alertsLoading: false })
    }
  },

  loadSplits: async (runId) => {
    try {
      const splits = await api.getSplits(runId)
      set({ splits: splits ?? [] })
    } catch {
      // non-fatal — splits panel will show empty
    }
  },

  loadChecks: async (runId) => {
    try {
      const checks = await api.getChecks(runId)
      set({ checks: checks ?? [] })
    } catch {
      // non-fatal
    }
  },

  confirmStep: async (runId, stepId) => {
    set({ error: null })
    try {
      await api.confirmStep(runId, stepId)
      await Promise.all([
        get().loadRunState(runId),
        get().loadAlerts(runId),
        get().loadChecks(runId),
      ])
    } catch (e) {
      set({ error: String(e) })
    }
  },

  skipStep: async (runId, stepId) => {
    set({ error: null })
    try {
      await api.skipStep(runId, stepId)
      await get().loadRunState(runId)
    } catch (e) {
      set({ error: String(e) })
    }
  },

  undoStep: async (runId, stepId) => {
    set({ error: null })
    try {
      await api.undoStep(runId, stepId)
      await Promise.all([
        get().loadRunState(runId),
        get().loadAlerts(runId),
      ])
    } catch (e) {
      set({ error: String(e) })
    }
  },

  finishRun: async (runId) => {
    set({ error: null })
    try {
      await api.finishRun(runId)
      await Promise.all([
        get().loadRunState(runId),
        get().loadSplits(runId),
      ])
    } catch (e) {
      set({ error: String(e) })
    }
  },

  abandonRun: async (runId) => {
    set({ error: null })
    try {
      await api.abandonRun(runId)
      await get().loadRunState(runId)
    } catch (e) {
      set({ error: String(e) })
    }
  },

  answerCheck: async (runId, checkId, value) => {
    set({ error: null })
    try {
      await api.answerCheck(runId, checkId, value)
      await get().loadChecks(runId)
    } catch (e) {
      set({ error: String(e) })
    }
  },

  setStepFilter: (patch) =>
    set((s) => ({ stepFilter: { ...s.stepFilter, ...patch } })),

  clearError: () => set({ error: null }),
})
)
