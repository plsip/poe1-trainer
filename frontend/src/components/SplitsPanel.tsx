import type { Split, GuideStep, RankingEntry } from '../api/types'

function formatMs(ms: number) {
  const totalSec = Math.floor(ms / 1000)
  const h = Math.floor(totalSec / 3600)
  const m = Math.floor((totalSec % 3600) / 60)
  const s = totalSec % 60
  return h > 0
    ? `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
    : `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

interface Props {
  splits: Split[]
  steps: GuideStep[]
  elapsedMs: number
  ranking: RankingEntry[]
}

export function SplitsPanel({ splits, steps, elapsedMs, ranking }: Props) {
  const stepMap = new Map(steps.map((s) => [s.id, s]))

  const pb = ranking.length > 0 ? ranking[0] : null
  const pbMs = pb?.total_ms ?? null

  return (
    <div className="panel">
      <h3 className="panel-title">Czasy etapów</h3>

      {pbMs !== null && (
        <div style={{ marginBottom: '0.5rem', fontSize: '0.82rem', color: '#aaa' }}>
          Rekord osobisty (PB):{' '}
          <strong style={{ color: '#ffd166' }}>{formatMs(pbMs)}</strong>
          {' · '}bieżący:{' '}
          <strong
            style={{ color: elapsedMs <= pbMs ? '#6ee7b7' : '#ff6b35' }}
          >
            {formatMs(elapsedMs)}
          </strong>
        </div>
      )}

      {splits.length === 0 ? (
        <p className="muted">Brak zarejestrowanych splitów.</p>
      ) : (
        <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.82rem' }}>
          <thead>
            <tr style={{ borderBottom: '1px solid #333' }}>
              <th style={{ textAlign: 'left', padding: '0.25rem 0.4rem', color: '#888', fontWeight: 500 }}>
                Krok
              </th>
              <th style={{ textAlign: 'right', padding: '0.25rem 0.4rem', color: '#888', fontWeight: 500 }}>
                Czas
              </th>
            </tr>
          </thead>
          <tbody>
            {splits.map((split) => {
              const step = stepMap.get(split.step_id)
              return (
                <tr key={split.id} style={{ borderBottom: '1px solid #1e1e1e' }}>
                  <td style={{ padding: '0.25rem 0.4rem', color: '#ccc' }}>
                    {step
                      ? `A${step.act} · ${step.title || `Krok ${step.step_number}`}`
                      : `Krok #${split.step_id}`}
                  </td>
                  <td
                    style={{
                      padding: '0.25rem 0.4rem',
                      textAlign: 'right',
                      fontFamily: 'monospace',
                      color: '#e0e0e0',
                    }}
                  >
                    {formatMs(split.split_ms)}
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      )}
    </div>
  )
}
