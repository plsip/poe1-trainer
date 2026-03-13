import type { Alert } from '../api/types'

const priorityColor: Record<Alert['priority'], string> = {
  high: '#ff6b35',
  medium: '#ffd166',
  low: '#a8dadc',
}

const priorityLabel: Record<Alert['priority'], string> = {
  high: 'wysokie',
  medium: 'średnie',
  low: 'niskie',
}

const kindIcon: Record<Alert['kind'], string> = {
  gem: '💎',
  gear: '⚔️',
}

interface Props {
  alerts: Alert[]
  loading?: boolean
}

export function AlertPanel({ alerts, loading }: Props) {
  if (loading) {
    return (
      <div className="panel">
        <h3 className="panel-title">Alerty</h3>
        <p className="muted">Ładowanie…</p>
      </div>
    )
  }

  const gems = alerts.filter((a) => a.kind === 'gem')
  const gear = alerts.filter((a) => a.kind === 'gear')

  if (alerts.length === 0) {
    return (
      <div className="panel">
        <h3 className="panel-title">Alerty</h3>
        <p className="muted">Brak alertów dla bieżącego kroku.</p>
      </div>
    )
  }

  return (
    <div className="panel">
      <h3 className="panel-title">Alerty</h3>

      {gems.length > 0 && (
        <section style={{ marginBottom: '0.75rem' }}>
          <div className="panel-section-label">💎 Gemy</div>
          <ul className="alert-list">
            {gems.map((a, i) => (
              <AlertItem key={i} alert={a} />
            ))}
          </ul>
        </section>
      )}

      {gear.length > 0 && (
        <section>
          <div className="panel-section-label">⚔️ Ekwipunek</div>
          <ul className="alert-list">
            {gear.map((a, i) => (
              <AlertItem key={i} alert={a} />
            ))}
          </ul>
        </section>
      )}
    </div>
  )
}

function AlertItem({ alert }: { alert: Alert }) {
  const color = priorityColor[alert.priority]
  return (
    <li
      style={{
        borderLeft: `3px solid ${color}`,
        padding: '0.4rem 0.6rem',
        marginBottom: '0.35rem',
        background: 'rgba(255,255,255,0.03)',
        borderRadius: '0 4px 4px 0',
      }}
    >
      <div style={{ display: 'flex', alignItems: 'baseline', gap: '0.5rem' }}>
        <span style={{ fontWeight: 600, fontSize: '0.9rem' }}>
          {kindIcon[alert.kind]} {alert.slot ? `[${alert.slot}] ` : ''}{alert.description}
        </span>
        <span style={{ fontSize: '0.7rem', color, flexShrink: 0 }}>
          {priorityLabel[alert.priority]}
        </span>
      </div>
      {alert.notes && (
        <div style={{ fontSize: '0.8rem', color: '#999', marginTop: '0.15rem' }}>
          {alert.notes}
        </div>
      )}
    </li>
  )
}
