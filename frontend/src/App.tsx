import { BrowserRouter, Routes, Route, Link } from 'react-router-dom'
import { HomePage } from './pages/HomePage'
import { GuidePage } from './pages/GuidePage/GuidePage'
import { RunPage } from './pages/RunPage/RunPage'
import { HistoryPage } from './pages/HistoryPage/HistoryPage'

export function App() {
  return (
    <BrowserRouter>
      <nav
        style={{
          padding: '0.6rem 1.5rem',
          borderBottom: '1px solid #333',
          display: 'flex',
          gap: '1.5rem',
          alignItems: 'center',
          background: '#111',
        }}
      >
        <Link to="/" style={{ fontWeight: 'bold', color: '#ffd166', textDecoration: 'none' }}>
          PoE1 Trainer
        </Link>
      </nav>

      <main style={{ minHeight: 'calc(100vh - 48px)', background: '#1a1a1a', color: '#e0e0e0' }}>
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/guides/:slug" element={<GuidePage />} />
          <Route path="/guides/:slug/history" element={<HistoryPage />} />
          <Route path="/runs/:runId" element={<RunPage />} />
        </Routes>
      </main>
    </BrowserRouter>
  )
}
