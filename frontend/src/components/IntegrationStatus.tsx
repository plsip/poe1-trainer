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

interface Props {
  /** ISO timestamp of the most recent area_entered event, if any. */
  lastAreaAt?: string
  /** Area name from the most recent logtail event, if any. */
  lastArea?: string
}

export function IntegrationStatus({ lastAreaAt, lastArea }: Props) {
  const logWatcherOnline = !!lastAreaAt
  const lastSeenLabel = lastAreaAt
    ? new Date(lastAreaAt).toLocaleTimeString('pl-PL', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
    : undefined

  return (
    <div className="panel">
      <h3 className="panel-title">Integracje</h3>

      <StatusDot
        online={logWatcherOnline}
        label="Log watcher"
        detail={
          lastArea
            ? `${lastArea} · ${lastSeenLabel}`
            : logWatcherOnline
            ? lastSeenLabel
            : 'offline'
        }
      />

      <StatusDot
        online={false}
        label="GGG API"
        detail="Faza 2"
      />
    </div>
  )
}
