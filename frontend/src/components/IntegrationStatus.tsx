import { useEffect, useState } from 'react'
import * as api from '../api/client'
import type { IntegrationStatus as IntegrationStatusType } from '../api/types'

/** Displays current integration status (log watcher, GGG API). */

interface StatusDotProps {
  online: boolean
  label: string
  detail?: string
}

function StatusDot({ online, label, detail }: StatusDotProps) {
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: '0.45rem',
        marginBottom: '0.3rem',
        fontSize: '0.82rem',
      }}
    >
      <span
        style={{
          width: 8,
          height: 8,
          borderRadius: '50%',
          background: online ? '#6ee7b7' : '#555',
          flexShrink: 0,
          boxShadow: online ? '0 0 4px #6ee7b7' : 'none',
        }}
      />
      <span style={{ color: '#ccc' }}>{label}</span>
      {detail && <span style={{ color: '#666', marginLeft: 'auto', fontSize: '0.75rem' }}>{detail}</span>}
    </div>
  )
}

const watcherLabel: Record<IntegrationStatusType['log_watcher'], string> = {
  active: 'aktywny',
  waiting_for_file: 'brak pliku',
  waiting_for_new_lines: 'czeka na dane',
  game_not_running: 'gra wyłączona',
  parser_error: 'błąd parsera',
  disabled: 'offline',
}

export function IntegrationStatus() {
  const [status, setStatus] = useState<IntegrationStatusType | null>(null)

  useEffect(() => {
    const load = () => api.getIntegrationStatus().then(setStatus).catch(() => {})
    load()
    const id = setInterval(load, 5_000)
    return () => clearInterval(id)
  }, [])

  const watcherStatus = status?.log_watcher ?? 'disabled'
  const watcherOnline = watcherStatus === 'active' || watcherStatus === 'waiting_for_new_lines'

  return (
    <div className="panel">
      <h3 className="panel-title">Integracje</h3>

      <StatusDot
        online={watcherOnline}
        label="Log watcher"
        detail={watcherLabel[watcherStatus]}
      />

      <StatusDot
        online={false}
        label="GGG API"
        detail="Faza 2"
      />
    </div>
  )
}
