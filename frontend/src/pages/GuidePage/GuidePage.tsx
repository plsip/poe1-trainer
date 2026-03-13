import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useAppStore } from '../../store/appStore'
import type { GuideStep } from '../../api/types'

export function GuidePage() {
  const { slug } = useParams<{ slug: string }>()
  const navigate = useNavigate()
  const { activeGuide, runState, loadGuide, startRun, error } = useAppStore()

  const [charName, setCharName] = useState('')
  const [starting, setStarting] = useState(false)

  useEffect(() => {
    if (slug) loadGuide(slug)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [slug])

  const confirmedSet = new Set(runState?.confirmed_step_ids ?? [])

  const handleStartRun = async () => {
    if (!activeGuide) return
    setStarting(true)
    await startRun(activeGuide.id, charName)
    setStarting(false)
    const newRunId = useAppStore.getState().activeRun?.id
    if (newRunId) navigate(`/runs/${newRunId}`)
  }

  const groupedByAct = (steps: GuideStep[]) =>
    steps.reduce<Record<number, GuideStep[]>>((acc, s) => {
      ;(acc[s.act] = acc[s.act] ?? []).push(s)
      return acc
    }, {})

  if (!activeGuide) {
    return <p>Ładowanie poradnika…{error && <span style={{ color: 'red' }}> {error}</span>}</p>
  }

  const acts = groupedByAct(activeGuide.steps ?? [])

  return (
    <div style={{ maxWidth: 860, margin: '0 auto', padding: '1.5rem' }}>
      <h1 style={{ marginBottom: 4 }}>{activeGuide.title}</h1>
      <p style={{ color: '#888', marginTop: 0 }}>
        {activeGuide.build_name} · v{activeGuide.version}
      </p>

      {/* Start run form */}
      <section style={{ display: 'flex', gap: '0.75rem', alignItems: 'center', marginBottom: '2rem', flexWrap: 'wrap' }}>
        <input
          type="text"
          placeholder="Nazwa postaci (opcjonalnie)"
          value={charName}
          onChange={(e) => setCharName(e.target.value)}
          style={{ padding: '0.5rem 0.75rem', minWidth: 220 }}
        />
        <button
          onClick={handleStartRun}
          disabled={starting}
          style={{ padding: '0.5rem 1rem', cursor: 'pointer', fontWeight: 'bold' }}
        >
          {starting ? 'Startuję…' : '▶ Rozpocznij run'}
        </button>
        {runState && (
          <button
            onClick={() => navigate(`/runs/${runState.run.id}`)}
            style={{ padding: '0.5rem 1rem', cursor: 'pointer' }}
          >
            Wróć do aktywnego runu #{runState.run.id}
          </button>
        )}
      </section>

      {/* Steps by act */}
      {Object.entries(acts).map(([act, steps]) => (
        <section key={act} style={{ marginBottom: '2rem' }}>
          <h2 style={{ color: '#ffd166', marginBottom: '0.5rem' }}>Akt {act}</h2>
          <ol style={{ paddingLeft: '1.2rem', margin: 0 }}>
            {steps.map((s) => {
              const done = confirmedSet.has(s.id)
              const isCurrent = s.id === runState?.current_step_id
              return (
                <li
                  key={s.id}
                  style={{
                    marginBottom: '0.35rem',
                    color: done ? '#6ee7b7' : isCurrent ? '#ffd166' : 'inherit',
                    fontWeight: isCurrent ? 'bold' : 'normal',
                  }}
                >
                  {done && '✓ '}
                  {isCurrent && '→ '}
                  <span dangerouslySetInnerHTML={{ __html: s.description }} />
                  {s.is_checkpoint && (
                    <span style={{ marginLeft: '0.4rem', fontSize: '0.75rem', color: '#ff6b35' }}>
                      [milestone]
                    </span>
                  )}
                  {s.gem_requirements && s.gem_requirements.length > 0 && (
                    <span style={{ marginLeft: '0.4rem', fontSize: '0.75rem', color: '#a8dadc' }}>
                      💎 {s.gem_requirements.map((g) => g.gem_name).join(', ')}
                    </span>
                  )}
                </li>
              )
            })}
          </ol>
        </section>
      ))}
    </div>
  )
}
