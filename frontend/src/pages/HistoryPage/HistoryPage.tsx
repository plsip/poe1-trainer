import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import * as api from '../../api/client'
import type { RunSession, RankingEntry } from '../../api/types'

function formatMs(ms: number) {
  const totalSec = Math.floor(ms / 1000)
  const h = Math.floor(totalSec / 3600)
  const m = Math.floor((totalSec % 3600) / 60)
  const s = totalSec % 60
  return h > 0
    ? `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
    : `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

export function HistoryPage() {
  const { slug } = useParams<{ slug: string }>()
  const [runs, setRuns] = useState<RunSession[]>([])
  const [ranking, setRanking] = useState<RankingEntry[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!slug) return
    setLoading(true)
    Promise.all([api.listRuns(slug), api.getRanking(slug)])
      .then(([r, rank]) => {
        setRuns(r ?? [])
        setRanking(rank ?? [])
      })
      .catch((e) => setError(String(e)))
      .finally(() => setLoading(false))
  }, [slug])

  if (loading) return <p>Ładowanie historii…</p>
  if (error) return <p style={{ color: 'red' }}>{error}</p>

  return (
    <div style={{ maxWidth: 860, margin: '0 auto', padding: '1.5rem' }}>
      <h1>Historia runów — {slug}</h1>

      {/* Ranking */}
      <section style={{ marginBottom: '2rem' }}>
        <h2 style={{ color: '#ffd166' }}>Lokalny ranking (top 20)</h2>
        {ranking.length === 0 ? (
          <p style={{ color: '#888' }}>Brak zakończonych runów.</p>
        ) : (
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ borderBottom: '1px solid #444' }}>
                <th style={{ textAlign: 'left', padding: '0.4rem 0.6rem' }}>#</th>
                <th style={{ textAlign: 'left', padding: '0.4rem 0.6rem' }}>Postać</th>
                <th style={{ textAlign: 'left', padding: '0.4rem 0.6rem' }}>Czas</th>
                <th style={{ textAlign: 'left', padding: '0.4rem 0.6rem' }}>Data</th>
              </tr>
            </thead>
            <tbody>
              {ranking.map((entry, i) => (
                <tr key={entry.run_id} style={{ borderBottom: '1px solid #2a2a2a' }}>
                  <td style={{ padding: '0.4rem 0.6rem', color: i === 0 ? '#ffd166' : 'inherit' }}>
                    {i + 1}
                  </td>
                  <td style={{ padding: '0.4rem 0.6rem' }}>
                    <Link to={`/runs/${entry.run_id}`}>{entry.character_name || `Run #${entry.run_id}`}</Link>
                  </td>
                  <td style={{ padding: '0.4rem 0.6rem', fontFamily: 'monospace' }}>
                    {formatMs(entry.total_ms)}
                  </td>
                  <td style={{ padding: '0.4rem 0.6rem', color: '#888', fontSize: '0.85rem' }}>
                    {new Date(entry.started_at).toLocaleDateString('pl-PL')}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>

      {/* All runs */}
      <section>
        <h2>Wszystkie runy</h2>
        {runs.length === 0 ? (
          <p style={{ color: '#888' }}>Brak runów.</p>
        ) : (
          <ul style={{ listStyle: 'none', padding: 0 }}>
            {runs.map((r) => (
              <li key={r.id} style={{ marginBottom: '0.5rem' }}>
                <Link to={`/runs/${r.id}`}>
                  Run #{r.id} — {r.character_name || 'bez nazwy'}
                </Link>
                <span style={{ marginLeft: '0.5rem', color: '#888', fontSize: '0.85rem' }}>
                  {r.is_active ? '🟢 aktywny' : '⬛ zakończony'} ·{' '}
                  {new Date(r.started_at).toLocaleString('pl-PL')}
                </span>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  )
}
