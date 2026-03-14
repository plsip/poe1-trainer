import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useAppStore } from '../../store/appStore'
import { formatGuideVersion } from '../../utils/guideVersion'

export function GuidePage() {
  const { slug } = useParams<{ slug: string }>()
  const navigate = useNavigate()
  const { activeGuide, runState, loadGuide, startRun, error } = useAppStore()

  const [charName, setCharName] = useState('')
  const [starting, setStarting] = useState(false)

  useEffect(() => {
    if (!slug) return
    loadGuide(slug)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [slug])

  const handleStartRun = async () => {
    if (!activeGuide) return
    setStarting(true)
    await startRun(activeGuide.id, charName)
    setStarting(false)
    const newRunId = useAppStore.getState().activeRun?.id
    if (newRunId) navigate(`/runs/${newRunId}`)
  }

  if (!activeGuide) {
    return (
      <div style={{ padding: '2rem', color: '#888' }}>
        Ładowanie poradnika…
        {error && <span style={{ color: '#ff6b35', marginLeft: '0.5rem' }}>{error}</span>}
      </div>
    )
  }

  const steps = activeGuide.steps ?? []
  const acts = [...new Set(steps.map((s) => s.act))].sort((a, b) => a - b)

  return (
    <div style={{ maxWidth: 900, margin: '0 auto', padding: '1.5rem' }}>
      {/* Guide header */}
      <h1 style={{ marginBottom: 4 }}>{activeGuide.title}</h1>
      <p style={{ color: '#888', marginTop: 0, marginBottom: '1.5rem' }}>
        {activeGuide.build_name} · rev: {formatGuideVersion(activeGuide.version)}
      </p>

      {/* Start run form */}
      <section
        style={{
          display: 'flex',
          gap: '0.75rem',
          alignItems: 'center',
          marginBottom: '2rem',
          flexWrap: 'wrap',
          padding: '1rem',
          background: 'rgba(255,209,102,0.05)',
          border: '1px solid rgba(255,209,102,0.15)',
          borderRadius: 6,
        }}
      >
        <input
          type="text"
          placeholder="Nazwa postaci (opcjonalnie)"
          value={charName}
          onChange={(e) => setCharName(e.target.value)}
          style={{ minWidth: 220 }}
          onKeyDown={(e) => e.key === 'Enter' && handleStartRun()}
        />
        <button
          className="btn-primary"
          onClick={handleStartRun}
          disabled={starting}
        >
          {starting ? 'Startuję…' : '▶ Rozpocznij run'}
        </button>
        {runState && (
          <button onClick={() => navigate(`/runs/${runState.run.id}`)}>
            Wróć do aktywnego runu #{runState.run.id}
          </button>
        )}
      </section>

      {/* Overview: acts summary */}
      <section style={{ marginBottom: '2rem' }}>
        <div
          style={{
            display: 'flex',
            gap: '0.5rem',
            flexWrap: 'wrap',
          }}
        >
          {acts.map((act) => {
            const actSteps = steps.filter((s) => s.act === act)
            const checkpoints = actSteps.filter((s) => s.is_checkpoint).length
            return (
              <div
                key={act}
                style={{
                  padding: '0.5rem 0.85rem',
                  background: 'rgba(255,255,255,0.04)',
                  border: '1px solid #2a2a2a',
                  borderRadius: 5,
                  fontSize: '0.82rem',
                }}
              >
                <div style={{ color: '#ffd166', fontWeight: 600 }}>Akt {act}</div>
                <div style={{ color: '#888' }}>
                  {actSteps.length} kroków · {checkpoints} ★
                </div>
              </div>
            )
          })}
        </div>
      </section>

    </div>
  )
}
