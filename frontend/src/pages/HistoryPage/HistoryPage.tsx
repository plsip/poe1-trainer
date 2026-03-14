import { useEffect, useState } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import * as api from '../../api/client'
import type { RunSession, DetailedRankingEntry, RankingStats } from '../../api/types'
import { getCanonicalGuideSlug } from '../../utils/guideSlug'

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
  const navigate = useNavigate()
  const [runs, setRuns] = useState<RunSession[]>([])
  const [ranking, setRanking] = useState<DetailedRankingEntry[]>([])
  const [stats, setStats] = useState<RankingStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const canonicalSlug = getCanonicalGuideSlug(slug)

  useEffect(() => {
    if (!slug || !canonicalSlug) return
    if (canonicalSlug !== slug) {
      navigate(`/guides/${canonicalSlug}/history`, { replace: true })
      return
    }
    setLoading(true)
    Promise.all([
      api.listRuns(canonicalSlug),
      api.getRanking(canonicalSlug),
      api.getRankingStats(canonicalSlug),
    ])
      .then(([r, rank, st]) => {
        setRuns(r ?? [])
        setRanking(rank ?? [])
        setStats(st ?? null)
      })
      .catch((e) => setError(String(e)))
      .finally(() => setLoading(false))
  }, [slug, canonicalSlug, navigate])

  if (loading) return <p>Ładowanie historii…</p>
  if (error) return <p style={{ color: 'red' }}>{error}</p>

  // Collect all unique act numbers from ranking data
  const actKeys = Array.from(
    new Set(ranking.flatMap((e) => Object.keys(e.act_splits))),
  ).sort((a, b) => Number(a) - Number(b))

  return (
    <div style={{ maxWidth: 960, margin: '0 auto', padding: '1.5rem' }}>
      <h1>Historia runów — {canonicalSlug}</h1>

      {/* ── Statystyki ── */}
      {stats && stats.count > 0 && (
        <section style={{ marginBottom: '1.5rem', display: 'flex', gap: '2rem', flexWrap: 'wrap' }}>
          <div>
            <div style={{ color: '#888', fontSize: '0.78rem' }}>Liczba zakończonych runów</div>
            <div style={{ fontSize: '1.3rem', fontWeight: 600 }}>{stats.count}</div>
          </div>
          {stats.pb_ms != null && (
            <div>
              <div style={{ color: '#888', fontSize: '0.78rem' }}>Rekord osobisty (PB)</div>
              <div style={{ fontSize: '1.3rem', fontWeight: 600, color: '#ffd166' }}>
                {formatMs(stats.pb_ms)}
              </div>
            </div>
          )}
          {stats.median_ms != null && (
            <div>
              <div style={{ color: '#888', fontSize: '0.78rem' }}>Mediana</div>
              <div style={{ fontSize: '1.3rem', fontWeight: 600, color: '#6ee7b7' }}>
                {formatMs(stats.median_ms)}
              </div>
            </div>
          )}
          {stats.last_run_ms != null && (
            <div>
              <div style={{ color: '#888', fontSize: '0.78rem' }}>Ostatni run</div>
              <div style={{ fontSize: '1.3rem', fontWeight: 600 }}>
                {formatMs(stats.last_run_ms)}
                {stats.last_run_id && (
                  <Link
                    to={`/runs/${stats.last_run_id}`}
                    style={{ marginLeft: '0.5rem', fontSize: '0.78rem', color: '#aaa' }}
                  >
                    (otwórz)
                  </Link>
                )}
              </div>
            </div>
          )}
        </section>
      )}

      {/* ── Ranking ── */}
      <section style={{ marginBottom: '2rem', overflowX: 'auto' }}>
        <h2 style={{ color: '#ffd166' }}>Lokalny ranking (top 20)</h2>
        {ranking.length === 0 ? (
          <p style={{ color: '#888' }}>Brak zakończonych runów z wyliczonym rankingiem.</p>
        ) : (
          <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.87rem' }}>
            <thead>
              <tr style={{ borderBottom: '1px solid #444' }}>
                <th style={{ textAlign: 'left', padding: '0.4rem 0.5rem' }}>#</th>
                <th style={{ textAlign: 'left', padding: '0.4rem 0.5rem' }}>Postać</th>
                <th style={{ textAlign: 'right', padding: '0.4rem 0.5rem' }}>Łączny czas</th>
                {actKeys.map((a) => (
                  <th key={a} style={{ textAlign: 'right', padding: '0.4rem 0.5rem', color: '#888' }}>
                    A{a}
                  </th>
                ))}
                <th style={{ textAlign: 'left', padding: '0.4rem 0.5rem', color: '#666' }}>Data</th>
              </tr>
            </thead>
            <tbody>
              {ranking.map((entry) => (
                <tr
                  key={entry.run_id}
                  style={{
                    borderBottom: '1px solid #2a2a2a',
                    background: entry.is_pb ? 'rgba(255, 209, 102, 0.05)' : 'transparent',
                  }}
                >
                  <td style={{ padding: '0.4rem 0.5rem', color: entry.is_pb ? '#ffd166' : '#888' }}>
                    {entry.rank}
                    {entry.is_pb && (
                      <span
                        title="Rekord osobisty"
                        style={{ marginLeft: '0.3rem', fontSize: '0.7rem', background: '#ffd166', color: '#111', borderRadius: 3, padding: '0 3px' }}
                      >
                        PB
                      </span>
                    )}
                  </td>
                  <td style={{ padding: '0.4rem 0.5rem' }}>
                    <Link to={`/runs/${entry.run_id}`}>
                      {entry.character_name || `Run #${entry.run_id}`}
                    </Link>
                  </td>
                  <td style={{ padding: '0.4rem 0.5rem', textAlign: 'right', fontFamily: 'monospace', color: entry.is_pb ? '#ffd166' : '#e0e0e0' }}>
                    {formatMs(entry.total_ms)}
                  </td>
                  {actKeys.map((a) => (
                    <td key={a} style={{ padding: '0.4rem 0.5rem', textAlign: 'right', fontFamily: 'monospace', color: '#aaa' }}>
                      {entry.act_splits[a] != null ? formatMs(entry.act_splits[a]) : '—'}
                    </td>
                  ))}
                  <td style={{ padding: '0.4rem 0.5rem', color: '#666', fontSize: '0.8rem' }}>
                    {new Date(entry.started_at).toLocaleDateString('pl-PL')}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>

      {/* ── Wszystkie runy ── */}
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
                  {r.is_active ? '🟢 aktywny' : r.status === 'abandoned' ? '🔴 porzucony' : '⬛ zakończony'}
                  {' · '}
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

