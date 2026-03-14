import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { useAppStore } from '../store/appStore'
import { formatGuideVersion } from '../utils/guideVersion'
import * as api from '../api/client'
import type { RunSession } from '../api/types'

export function HomePage() {
  const { guides, guidesLoading, error, loadGuides } = useAppStore()
  const [activeRuns, setActiveRuns] = useState<RunSession[]>([])

  useEffect(() => {
    loadGuides()
    api.listActiveRuns().then(setActiveRuns).catch(() => {})
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <div style={{ maxWidth: 640, margin: '0 auto', padding: '2rem' }}>
      <h1>PoE1 Trainer</h1>
      <p style={{ color: '#aaa' }}>Wybierz poradnik, żeby rozpocząć naukę.</p>

      {activeRuns.length > 0 && (
        <div style={{
          background: '#1a2a1a',
          border: '1px solid #6ee7b7',
          borderRadius: 8,
          padding: '1rem 1.25rem',
          marginBottom: '1.5rem',
        }}>
          <div style={{ fontSize: '0.8rem', color: '#6ee7b7', marginBottom: '0.5rem', textTransform: 'uppercase', letterSpacing: 1 }}>
            Aktywny run
          </div>
          {activeRuns.map((run) => {
            const guide = guides.find((g) => g.id === run.guide_id)
            const isPaused = !!run.paused_at
            return (
              <div key={run.id} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: '1rem' }}>
                <div>
                  <span style={{ fontWeight: 600, color: '#fff' }}>
                    {run.character_name || 'Bez nazwy'}
                  </span>
                  {guide && (
                    <span style={{ color: '#888', marginLeft: '0.5rem', fontSize: '0.9rem' }}>
                      · {guide.title}
                    </span>
                  )}
                  {isPaused && (
                    <span style={{ color: '#ffd166', marginLeft: '0.5rem', fontSize: '0.85rem' }}>⏸ zapauzowany</span>
                  )}
                </div>
                <Link
                  to={`/runs/${run.id}`}
                  style={{
                    padding: '0.35rem 0.9rem',
                    background: '#6ee7b7',
                    color: '#0d1117',
                    borderRadius: 6,
                    fontWeight: 700,
                    textDecoration: 'none',
                    fontSize: '0.9rem',
                    whiteSpace: 'nowrap',
                  }}
                >
                  Wróć do runu →
                </Link>
              </div>
            )
          })}
        </div>
      )}

      {guidesLoading && <p>Ładowanie poradników…</p>}
      {error && <p style={{ color: 'red' }}>{error}</p>}

      <ul style={{ listStyle: 'none', padding: 0 }}>
        {guides.map((g) => (
          <li key={g.id} style={{ marginBottom: '1rem' }}>
            <Link
              to={`/guides/${g.slug}`}
              style={{ fontSize: '1.15rem', textDecoration: 'none', color: '#ffd166' }}
            >
              {g.title}
            </Link>
            <div style={{ fontSize: '0.85rem', color: '#888', marginTop: 2 }}>
              {g.build_name} · rev: {formatGuideVersion(g.version)}
              {' · '}
              <Link to={`/guides/${g.slug}/history`} style={{ color: '#a8dadc' }}>
                historia i ranking
              </Link>
            </div>
          </li>
        ))}
        {!guidesLoading && guides.length === 0 && !error && (
          <li style={{ color: '#888' }}>
            Brak poradników. Zaimportuj guide przez backend.
          </li>
        )}
      </ul>
    </div>
  )
}
