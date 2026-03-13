import { create } from 'zustand'
import type { Guide, CurrentState, Recommendation, RunSession } from '../api/types'
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

  // errors
  error: string | null

  // actions
  loadGuides: () => Promise<void>
  loadGuide: (slug: string) => Promise<void>
  startRun: (guideId: number, characterName: string) => Promise<void>
  loadRunState: (runId: number) => Promise<void>
  confirmStep: (runId: number, stepId: number) => Promise<void>
  finishRun: (runId: number) => Promise<void>
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
      const [state, recs] = await Promise.all([
        api.getRunState(runId),
        api.getRecommendations(runId),
      ])
      set({ runState: state, recommendations: recs, stateLoading: false })
    } catch (e) {
      set({ stateLoading: false, error: String(e) })
    }
  },

  confirmStep: async (runId, stepId) => {
    set({ error: null })
    try {
      await api.confirmStep(runId, stepId)
      await get().loadRunState(runId)
    } catch (e) {
      set({ error: String(e) })
    }
  },

  finishRun: async (runId) => {
    set({ error: null })
    try {
      await api.finishRun(runId)
      await get().loadRunState(runId)
    } catch (e) {
      set({ error: String(e) })
    }
  },

  clearError: () => set({ error: null }),
}))
