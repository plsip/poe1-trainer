import { useEffect, useRef } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useAppStore } from '../../store/appStore'
import { RunTimer } from '../../components/RunTimer'
import { StepList } from '../../components/StepList'
import { ChecksPanel } from '../../components/ChecksPanel'
import { LogTailPanel } from '../../components/LogTailPanel'
import { ReplayLogButton } from '../../components/ReplayLogButton'
import type { GuideStep, CurrentState } from '../../api/types'
import * as api from '../../api/client'

export function RunPage() {
  const { runId } = useParams<{ runId: string }>()
  const navigate = useNavigate()

  const {
    runState,
    stateLoading,
    error,
    checks,
    stepFilter,
    activeGuide,
    loadRunState,
    loadSplits,
    loadChecks,
    confirmStep,
    confirmAct,
    skipStep,
    undoStep,
    finishRun,
    abandonRun,
    answerCheck,
    setStepFilter,
  } = useAppStore()

  const id = Number(runId)
  const isActive = runState?.run.is_active ?? false
  const isPaused = !!runState?.run.paused_at
  const steps: GuideStep[] = activeGuide?.steps ?? []

  // Initial load
  useEffect(() => {
    if (!id) return
    loadRunState(id)
    loadSplits(id)
    loadChecks(id)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id])

  // SSE stream — replaces polling; backend pushes run state updates
  const sseRef = useRef<EventSource | null>(null)
  useEffect(() => {
    if (!id) return
    sseRef.current?.close()
    const es = api.subscribeToRunStream(id)
    sseRef.current = es

    es.addEventListener('state', (e) => {
      try {
        const state: CurrentState = JSON.parse((e as MessageEvent).data)
        useAppStore.setState({ runState: state })
        // Reload checks when the step changes
        loadChecks(id)
      } catch {
        // ignore malformed events
      }
    })

    es.onerror = () => {
      // EventSource reconnects automatically; nothing to do here
    }

    return () => {
      es.close()
      sseRef.current = null
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id])

  // ─── Action handlers ─────────────────────────────────────────────────────

  const handleConfirm = (stepId: number) => confirmStep(id, stepId)
  const handleConfirmAct = (act: number) => confirmAct(id, act)
  const handleSkip = (stepId: number) => skipStep(id, stepId)
  const handleUndo = (stepId: number) => undoStep(id, stepId)
  const handleAnswer = (checkId: number, value: string) => answerCheck(id, checkId, value)

  const handleFinish = async () => {
    if (!confirm('Zakończyć run? Zostaną zapisane finalne splity.')) return
    await finishRun(id)
    loadSplits(id)
  }

  const handleAbandon = async () => {
    if (!confirm('Porzucić run? Tego nie można cofnąć.')) return
    await abandonRun(id)
    navigate(-1)
  }

  const handlePause = async () => {
    if (isPaused) {
      await api.resumeRun(id).catch(() => {})
    } else {
      await api.pauseRun(id).catch(() => {})
    }
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
            isActive={isActive && !isPaused}
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
                  title={isPaused ? 'Wznów timer' : 'Wstrzymaj timer (AFK)'}
                  style={{ color: isPaused ? '#ffd166' : undefined }}
                >
                  {isPaused ? '▶ Wznów' : '⏸ Pauza'}
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
                <ReplayLogButton runId={id} onDone={() => loadRunState(id)} />
                {isPaused && (
                  <button className="btn-sm" onClick={() => navigate(-1)}>
                    ← Wróć
                  </button>
                )}
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

      {/* ── 3-column layout: steps | logtail | sidebar ── */}
      <div className="run-columns">
        {/* Left: step list (narrower) */}
        <div className="run-col-steps">
          <StepList
            steps={steps}
            state={runState}
            filter={stepFilter}
            isActive={isActive}
            onFilterChange={setStepFilter}
            onConfirm={handleConfirm}
            onConfirmAct={handleConfirmAct}
            onSkip={handleSkip}
            onUndo={handleUndo}
          />
        </div>

        {/* Right: Client.txt log + checks */}
        <div className="run-col-logtail">
          <div className="run-sidebar-sticky">
            <LogTailPanel isActive={isActive} />
            {checks.some((c) => !c.is_confirmed) && (
              <ChecksPanel checks={checks} onAnswer={handleAnswer} />
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

