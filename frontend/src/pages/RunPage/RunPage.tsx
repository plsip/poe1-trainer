import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useAppStore } from '../../store/appStore'
import { RecommendationList } from '../../components/RecommendationList'
import { RunTimer } from '../../components/RunTimer'
import type { GuideStep } from '../../api/types'

export function RunPage() {
  const { runId } = useParams<{ runId: string }>()
  const navigate = useNavigate()
  const { runState, recommendations, stateLoading, error, loadRunState, confirmStep, finishRun, activeGuide } =
    useAppStore()

  const [confirming, setConfirming] = useState(false)

  const id = Number(runId)

  useEffect(() => {
    if (id) loadRunState(id)
    // Refresh state every 10 seconds when active.
    const interval = setInterval(() => {
      if (id && runState?.run.is_active) loadRunState(id)
    }, 10_000)
    return () => clearInterval(interval)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id])

  const currentStep: GuideStep | undefined = activeGuide?.steps?.find(
    (s) => s.id === runState?.current_step_id,
  )

  const handleConfirm = async () => {
    if (!runState?.current_step_id) return
    setConfirming(true)
    await confirmStep(id, runState.current_step_id)
    setConfirming(false)
  }

  const handleFinish = async () => {
    if (!confirm('Zakończyć run?')) return
    await finishRun(id)
  }

  if (stateLoading && !runState) {
    return <p>Ładowanie stanu runu…</p>
  }
  if (error) {
    return <p style={{ color: 'red' }}>{error}</p>
  }
  if (!runState) {
    return <p>Nie znaleziono runu.</p>
  }

  const confirmedCount = runState.confirmed_step_ids.length
  const totalSteps = activeGuide?.steps?.length ?? 0

  return (
    <div style={{ maxWidth: 760, margin: '0 auto', padding: '1.5rem' }}>
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
        <div>
          <h1 style={{ margin: 0, fontSize: '1.4rem' }}>
            {runState.run.character_name || 'Bez nazwy'} — Run #{runState.run.id}
          </h1>
          <small style={{ color: '#888' }}>
            {confirmedCount}/{totalSteps} kroków · {runState.run.is_active ? 'aktywny' : 'zakończony'}
          </small>
        </div>
        <RunTimer
          startedAt={runState.run.started_at}
          isActive={runState.run.is_active}
          serverElapsedMs={runState.elapsed_ms}
        />
      </div>

      {/* Current step */}
      {currentStep ? (
        <section style={{ background: 'rgba(255,200,100,0.08)', border: '1px solid #8a6a20', borderRadius: 6, padding: '1rem', marginBottom: '1.2rem' }}>
          <h2 style={{ margin: '0 0 0.4rem', fontSize: '1.1rem', color: '#ffd166' }}>
            Bieżący krok — Akt {currentStep.act}, krok {currentStep.step_number}
          </h2>
          <p style={{ margin: 0 }} dangerouslySetInnerHTML={{ __html: currentStep.description }} />
          {currentStep.is_checkpoint && (
            <span style={{ display: 'inline-block', marginTop: '0.5rem', background: '#ff6b35', color: '#fff', borderRadius: 4, padding: '0.15rem 0.5rem', fontSize: '0.8rem' }}>
              Kamień milowy — wymagane potwierdzenie
            </span>
          )}
        </section>
      ) : (
        <section style={{ marginBottom: '1.2rem' }}>
          <p style={{ color: '#6ee7b7' }}>✓ Wszystkie kroki potwierdzone!</p>
        </section>
      )}

      {/* Actions */}
      {runState.run.is_active && (
        <div style={{ display: 'flex', gap: '0.75rem', marginBottom: '1.5rem' }}>
          <button
            onClick={handleConfirm}
            disabled={confirming || !runState.current_step_id}
            style={{ padding: '0.6rem 1.2rem', cursor: 'pointer', fontWeight: 'bold' }}
          >
            {confirming ? 'Potwierdzanie…' : '✓ Potwierdź krok'}
          </button>
          <button
            onClick={handleFinish}
            style={{ padding: '0.6rem 1.2rem', cursor: 'pointer', color: '#ff6b35', border: '1px solid #ff6b35', background: 'transparent' }}
          >
            Zakończ run
          </button>
          <button onClick={() => navigate(-1)} style={{ padding: '0.6rem 1.2rem', cursor: 'pointer' }}>
            ← Poradnik
          </button>
        </div>
      )}

      {/* Recommendations */}
      <section>
        <h2 style={{ fontSize: '1rem', marginBottom: '0.6rem', color: '#a8dadc' }}>Rekomendacje</h2>
        <RecommendationList recommendations={recommendations} />
      </section>
    </div>
  )
}
