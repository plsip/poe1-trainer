import { useState } from 'react'
import * as api from '../api/client'
import type { ReplayLogResult } from '../api/types'

interface Props {
  runId: number
  onDone?: () => void
}

/**
 * Narzędzie QA: wczytuje Client.txt od początku do aktywnego runa.
 * Pozwala zweryfikować czy splity pojawiają się przy właściwych krokach.
 */
export function ReplayLogButton({ runId, onDone }: Props) {
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<ReplayLogResult | null>(null)
  const [error, setError] = useState<string | null>(null)

  const handleReplay = async () => {
    if (
      !confirm(
        'Wczytać Client.txt do bieżącego runa?\n\n' +
          'Cały plik zostanie ponownie przetworzony od początku — ' +
          'zdarzenia (wejścia do stref, poziomy) zostaną dodane do runa.',
      )
    )
      return

    setLoading(true)
    setResult(null)
    setError(null)
    try {
      const res = await api.replayLog(runId)
      setResult(res)
      onDone?.()
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ display: 'inline-flex', flexDirection: 'column', gap: '0.25rem' }}>
      <button
        className="btn-sm"
        onClick={handleReplay}
        disabled={loading}
        title="Wczytaj Client.txt od początku do tego runa (narzędzie QA)"
        style={{ color: '#ffd166' }}
      >
        {loading ? '⏳ Wczytywanie…' : '⏩ Wczytaj log'}
      </button>

      {result && (
        <span
          style={{
            fontSize: '0.7rem',
            color: '#6ee7b7',
            whiteSpace: 'nowrap',
          }}
          title={`Czas: ${result.duration_ms} ms${result.parse_errors ? ` · błędy parsera: ${result.parse_errors}` : ''}`}
        >
          ✓ {result.events_dispatched} zd. / {result.lines_read} linii
        </span>
      )}

      {error && (
        <span style={{ fontSize: '0.7rem', color: '#ff6b35', whiteSpace: 'nowrap' }}>
          ✗ {error}
        </span>
      )}
    </div>
  )
}
