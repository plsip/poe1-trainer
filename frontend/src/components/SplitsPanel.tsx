import type { Split, GuideStep, DetailedRankingEntry, RunDeltasResponse } from '../api/types'

function formatMs(ms: number) {
  const totalSec = Math.floor(ms / 1000)
  const h = Math.floor(totalSec / 3600)
  const m = Math.floor((totalSec % 3600) / 60)
  const s = totalSec % 60
  return h > 0
    ? `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
    : `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

function formatDelta(delta: number | undefined) {
  if (delta === undefined || delta === null) return null
  const abs = Math.abs(delta)
  const sign = delta >= 0 ? '+' : '-'
  const totalSec = Math.floor(abs / 1000)
  const m = Math.floor(totalSec / 60)
  const s = totalSec % 60
  const ms = abs % 1000
  if (m > 0) return `${sign}${m}:${String(s).padStart(2, '0')}`
  if (s > 0) return `${sign}${s}.${String(Math.floor(ms / 100))}s`
  return `${sign}${ms}ms`
}

interface Props {
  splits: Split[]
  steps: GuideStep[]
  elapsedMs: number
  ranking: DetailedRankingEntry[]
  deltas?: RunDeltasResponse
}

export function SplitsPanel({ splits, steps, elapsedMs, ranking, deltas }: Props) {
  const stepMap = new Map(steps.map((s) => [s.id, s]))

  const pb = ranking.length > 0 ? ranking[0] : null
  const pbMs = pb?.total_ms ?? null

  // Build lookup map for deltas by step_id
  const deltaMap = new Map(
    (deltas?.splits ?? []).map((d) => [d.step_id, d]),
  )

  const showPBDelta = (deltas?.pb_run_id ?? 0) > 0 || ranking.length > 1
  const showPrevDelta = (deltas?.prev_run_id ?? 0) > 0

  return (
    <div className="panel">
      <h3 className="panel-title">Czasy etapów</h3>

      {pbMs !== null && (
        <div style={{ marginBottom: '0.5rem', fontSize: '0.82rem', color: '#aaa' }}>
          Rekord osobisty:{' '}
          <strong style={{ color: '#ffd166' }}>{formatMs(pbMs)}</strong>
          {' · '}bieżący:{' '}
          <strong style={{ color: elapsedMs <= pbMs ? '#6ee7b7' : '#ff6b35' }}>
            {formatMs(elapsedMs)}
          </strong>
          {pbMs > 0 && (
            <>
              {' · '}delta:{' '}
              <strong style={{ color: elapsedMs - pbMs <= 0 ? '#6ee7b7' : '#ff6b35' }}>
                {formatDelta(elapsedMs - pbMs)}
              </strong>
            </>
          )}
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
              {showPBDelta && (
                <th style={{ textAlign: 'right', padding: '0.25rem 0.4rem', color: '#ffd166', fontWeight: 500 }}>
                  vs PB
                </th>
              )}
              {showPrevDelta && (
                <th style={{ textAlign: 'right', padding: '0.25rem 0.4rem', color: '#aaa', fontWeight: 500 }}>
                  vs poprzedni
                </th>
              )}
            </tr>
          </thead>
          <tbody>
            {splits.map((split) => {
              const step = stepMap.get(split.step_id)
              const d = deltaMap.get(split.step_id)
              return (
                <tr key={split.id} style={{ borderBottom: '1px solid #1e1e1e' }}>
                  <td style={{ padding: '0.25rem 0.4rem', color: '#ccc' }}>
                    {step
                      ? `A${step.act} · ${step.title || `Krok ${step.step_number}`}`
                      : `Krok #${split.step_id}`}
                  </td>
                  <td style={{ padding: '0.25rem 0.4rem', textAlign: 'right', fontFamily: 'monospace', color: '#e0e0e0' }}>
                    {formatMs(split.split_ms)}
                  </td>
                  {showPBDelta && (
                    <td style={{
                      padding: '0.25rem 0.4rem',
                      textAlign: 'right',
                      fontFamily: 'monospace',
                      color: d?.delta_pb_ms != null
                        ? (d.delta_pb_ms <= 0 ? '#6ee7b7' : '#ff6b35')
                        : '#555',
                    }}>
                      {d?.delta_pb_ms != null ? formatDelta(d.delta_pb_ms) : '—'}
                    </td>
                  )}
                  {showPrevDelta && (
                    <td style={{
                      padding: '0.25rem 0.4rem',
                      textAlign: 'right',
                      fontFamily: 'monospace',
                      color: d?.delta_prev_ms != null
                        ? (d.delta_prev_ms <= 0 ? '#6ee7b7' : '#ff6b35')
                        : '#555',
                    }}>
                      {d?.delta_prev_ms != null ? formatDelta(d.delta_prev_ms) : '—'}
                    </td>
                  )}
                </tr>
              )
            })}
          </tbody>
        </table>
      )}
    </div>
  )
}
