import { useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useAppStore } from '../store/appStore'

export function HomePage() {
  const { guides, guidesLoading, error, loadGuides } = useAppStore()

  useEffect(() => {
    loadGuides()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <div style={{ maxWidth: 640, margin: '0 auto', padding: '2rem' }}>
      <h1>PoE1 Trainer</h1>
      <p style={{ color: '#aaa' }}>Wybierz poradnik, żeby rozpocząć naukę.</p>

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
              {g.build_name} · v{g.version}
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
