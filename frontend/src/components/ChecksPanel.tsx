import { useState } from 'react'
import type { ManualCheck } from '../api/types'

const checkTypeLabel: Record<string, string> = {
  gear: 'Ekwipunek',
  gem: 'Gem',
  level: 'Poziom',
  resist: 'Resistancje',
  flask: 'Flaszki',
  quest: 'Quest',
  free_form: 'Pytanie',
}

interface Props {
  checks: ManualCheck[]
  onAnswer: (checkId: number, value: string) => void
}

export function ChecksPanel({ checks, onAnswer }: Props) {
  const pending = checks.filter((c) => !c.is_confirmed)

  if (pending.length === 0) {
    return (
      <div className="panel">
        <h3 className="panel-title">Pytania kontrolne</h3>
        <p className="muted">Brak aktywnych pytań.</p>
      </div>
    )
  }

  return (
    <div className="panel">
      <h3 className="panel-title">Pytania kontrolne</h3>
      <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
        {pending.map((check) => (
          <CheckItem key={check.id} check={check} onAnswer={onAnswer} />
        ))}
      </ul>
    </div>
  )
}

function CheckItem({
  check,
  onAnswer,
}: {
  check: ManualCheck
  onAnswer: (id: number, value: string) => void
}) {
  const [value, setValue] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const handleSubmit = async () => {
    if (!value.trim()) return
    setSubmitting(true)
    await onAnswer(check.id, value.trim())
    setSubmitting(false)
  }

  return (
    <li
      style={{
        marginBottom: '0.75rem',
        padding: '0.6rem 0.75rem',
        background: 'rgba(255,255,255,0.03)',
        borderRadius: 4,
        border: '1px solid #2d3748',
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', marginBottom: '0.4rem' }}>
        <span
          style={{
            fontSize: '0.7rem',
            background: '#2a3a4a',
            color: '#a8dadc',
            borderRadius: 3,
            padding: '0.1rem 0.35rem',
          }}
        >
          {checkTypeLabel[check.check_type] ?? check.check_type}
        </span>
      </div>

      <p style={{ margin: '0 0 0.5rem', fontSize: '0.875rem', color: '#e0e0e0' }}>
        {check.prompt}
      </p>

      <div style={{ display: 'flex', gap: '0.4rem' }}>
        <input
          type="text"
          value={value}
          onChange={(e) => setValue(e.target.value)}
          placeholder="Twoja odpowiedź…"
          style={{ flex: 1, fontSize: '0.85rem' }}
          onKeyDown={(e) => e.key === 'Enter' && handleSubmit()}
        />
        <button
          className="btn-primary btn-sm"
          onClick={handleSubmit}
          disabled={submitting || !value.trim()}
        >
          {submitting ? '…' : 'Potwierdź'}
        </button>
        <button
          className="btn-sm"
          onClick={() => onAnswer(check.id, 'skip')}
          disabled={submitting}
        >
          Pomiń
        </button>
      </div>
    </li>
  )
}
