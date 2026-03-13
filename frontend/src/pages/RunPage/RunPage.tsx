import { useEffect, useRef, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useAppStore } from '../../store/appStore'
import { RecommendationList } from '../../components/RecommendationList'
import { RunTimer } from '../../components/RunTimer'
import { AlertPanel } from '../../components/AlertPanel'
import { StepList } from '../../components/StepList'
import { SplitsPanel } from '../../components/SplitsPanel'
import { IntegrationStatus } from '../../components/IntegrationStatus'
import { ChecksPanel } from '../../components/ChecksPanel'
import type { GuideStep, DetailedRankingEntry, RunDeltasResponse } from '../../api/types'
import * as api from '../../api/client'

export function RunPage() {
  const { runId } = useParams<{ runId: string }>()
  const navigate = useNavigate()

  const {
    runState,
    recommendations,
    stateLoading,
    error,
    alerts,
    alertsLoading,
    splits,
    checks,
    stepFilter,
    activeGuide,
    loadRunState,
    loadAlerts,
    loadSplits,
    loadChecks,
    confirmStep,
    skipStep,
    undoStep,
    finishRun,
    abandonRun,
    answerCheck,
    setStepFilter,
  } = useAppStore()

  const [ranking, setRanking] = useState<DetailedRankingEntry[]>([])
  const [deltas, setDeltas] = useState<RunDeltasResponse | undefined>()
  const [paused, setPaused] = useState(false)
  const [lastArea, setLastArea] = useState<string | undefined>()
  const [lastAreaAt, setLastAreaAt] = useState<string | undefined>()

  const id = Number(runId)
  const isActive = runState?.run.is_active ?? false
  const steps: GuideStep[] = activeGuide?.steps ?? []

  // Initial load
  useEffect(() => {
    if (!id) return
    loadRunState(id)
    loadAlerts(id)
    loadSplits(id)
    loadChecks(id)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id])

  // Load ranking and events once we have the guide slug
  const guideSlug = activeGuide?.slug
  useEffect(() => {
    if (!guideSlug) return
    api
      .getRanking(guideSlug)
      .then((r) => setRanking(r ?? []))
      .catch(() => {})
  }, [guideSlug])

  // Load split deltas whenever splits or run changes
  useEffect(() => {
    if (!id) return
    api
      .getSplitDeltas(id)
      .then((d) => setDeltas(d))
      .catch(() => {})
  }, [id])

  useEffect(() => {
    if (!id) return
    api
      .listEvents(id)
      .then((events) => {
        const areaEvents = events
          .filter((e) => e.event_type === 'area_entered')
          .sort((a, b) => new Date(b.occurred_at).getTime() - new Date(a.occurred_at).getTime())
        if (areaEvents.length > 0) {
          setLastArea(areaEvents[0].payload?.area)
          setLastAreaAt(areaEvents[0].occurred_at)
        }
      })
      .catch(() => {})
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id])

  // Periodic refresh when active
  const pollingRef = useRef<ReturnType<typeof setInterval> | null>(null)
  useEffect(() => {
    if (pollingRef.current) clearInterval(pollingRef.current)
    if (!id || !isActive) return
    pollingRef.current = setInterval(() => {
      loadRunState(id)
      loadAlerts(id)
      loadChecks(id)
    }, 10_000)
    return () => {
      if (pollingRef.current) clearInterval(pollingRef.current)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id, isActive])

  // ─── Action handlers ─────────────────────────────────────────────────────

  const handleConfirm = (stepId: number) => confirmStep(id, stepId)
  const handleSkip = (stepId: number) => skipStep(id, stepId)
  const handleUndo = (stepId: number) => undoStep(id, stepId)
  const handleAnswer = (checkId: number, value: string) => answerCheck(id, checkId, value)

  const handleFinish = async () => {
    if (!confirm('Zakończyć run? Zostaną zapisane finalne splity.')) return
    await finishRun(id)
    loadSplits(id)
    api.getSplitDeltas(id).then(setDeltas).catch(() => {})
  }

  const handleAbandon = async () => {
    if (!confirm('Porzucić run? Tego nie można cofnąć.')) return
    await abandonRun(id)
    navigate(-1)
  }

  const handlePause = async () => {
    if (paused) {
      await api.resumeRun(id).catch(() => {})
    } else {
      await api.pauseRun(id).catch(() => {})
    }
    setPaused(!paused)
    loadRunState(id)
  }

  // ─── Loading / error states ───────────────────────────────────────────────

  if (stateLoading && !runState) {
    return (
      <div style={{ padding: '2rem', color: '#888' }}>
        Ładowanie stanu runu…
      </div>
    )
  }
  if (error) {
    return (
      <div style={{ padding: '2rem' }}>
        <p style={{ color: '#ff6b35' }}>{error}</p>
        <button onClick={() => loadRunState(id)}>Ponów</button>
      </div>
    )
  }
  if (!runState) {
    return (
      <div style={{ padding: '2rem', color: '#888' }}>
        Nie znaleziono runu #{id}.
      </div>
    )
  }

  const confirmedCount = runState.confirmed_step_ids.length
  const totalSteps = steps.length

  // ─── Render ───────────────────────────────────────────────────────────────

  return (
    <div className="run-dashboard">
      {/* ── Top bar ── */}
      <header className="run-header">
        <div>
          <h1 className="run-title">
            {runState.run.character_name || 'Bez nazwy'}
            <span className="run-id"> · Run #{runState.run.id}</span>
          </h1>
          <div className="run-meta">
            {confirmedCount}/{totalSteps} kroków
            {' · '}
            <span
              style={{ color: isActive ? '#6ee7b7' : '#888' }}
            >
              {isActive ? 'aktywny' : runState.run.status}
            </span>
            {activeGuide && (
              <>
                {' · '}
                <span style={{ color: '#ffd166' }}>{activeGuide.title}</span>
              </>
            )}
          </div>
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
          <RunTimer
            startedAt={runState.run.started_at}
            isActive={isActive}
            serverElapsedMs={runState.elapsed_ms}
          />

          <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
            {isActive && (
              <>
                <button
                  className="btn-primary btn-sm"
                  onClick={() => runState.current_step_id && handleConfirm(runState.current_step_id)}
                  disabled={!runState.current_step_id}
                >
                  ✓ Potwierdź bieżący krok
                </button>
                <button
                  className="btn-sm"
                  onClick={handlePause}
                  title={paused ? 'Wznów timer' : 'Wstrzymaj timer (AFK)'}
                  style={{ color: paused ? '#ffd166' : undefined }}
                >
                  {paused ? '▶ Wznów' : '⏸ Pauza'}
                </button>
                <button
                  className="btn-sm"
                  onClick={handleFinish}
                >
                  Zakończ run
                </button>
                <button
                  className="btn-sm btn-danger"
                  onClick={handleAbandon}
                >
                  Porzuć
                </button>
              </>
            )}
            {!isActive && (
              <button className="btn-sm" onClick={() => navigate(-1)}>
                ← Wróć
              </button>
            )}
          </div>
        </div>
      </header>

      {/* ── 2-column layout ── */}
      <div className="run-columns">
        {/* Left: step list */}
        <div className="run-col-main">
          <StepList
            steps={steps}
            state={runState}
            filter={stepFilter}
            isActive={isActive}
            onFilterChange={setStepFilter}
            onConfirm={handleConfirm}
            onSkip={handleSkip}
            onUndo={handleUndo}
          />
        </div>

        {/* Right: sidebar panels */}
        <div className="run-col-sidebar">
          <AlertPanel alerts={alerts} loading={alertsLoading} />

          {checks.some((c) => !c.is_confirmed) && (
            <ChecksPanel checks={checks} onAnswer={handleAnswer} />
          )}

          <section className="panel">
            <h3 className="panel-title">Rekomendacje</h3>
            <RecommendationList recommendations={recommendations} />
          </section>

          <SplitsPanel
            splits={splits}
            steps={steps}
            elapsedMs={runState.elapsed_ms}
            ranking={ranking}
            deltas={deltas}
          />

          <IntegrationStatus lastArea={lastArea} lastAreaAt={lastAreaAt} />
        </div>
      </div>
    </div>
  )
}
