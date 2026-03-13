import { useState } from 'react'
import type { GuideStep, CurrentState, StepFilter, StepProgressStatus } from '../api/types'
import { StatusBadge } from './StatusBadge'

function deriveStatus(step: GuideStep, state: CurrentState | null): StepProgressStatus {
  if (!state) return 'pending'
  if (state.confirmed_step_ids.includes(step.id)) return 'completed'
  if (step.id === state.current_step_id) {
    return step.requires_manual ? 'needs_confirmation' : 'in_progress'
  }
  return 'pending'
}

interface Props {
  steps: GuideStep[]
  state: CurrentState | null
  filter: StepFilter
  isActive: boolean
  onFilterChange: (patch: Partial<StepFilter>) => void
  onConfirm: (stepId: number) => void
  onSkip: (stepId: number) => void
  onUndo: (stepId: number) => void
}

export function StepList({
  steps,
  state,
  filter,
  isActive,
  onFilterChange,
  onConfirm,
  onSkip,
  onUndo,
}: Props) {
  const [expandedId, setExpandedId] = useState<number | null>(null)
  const acts = [...new Set(steps.map((s) => s.act))].sort((a, b) => a - b)

  const filtered = steps.filter((s) => {
    if (filter.act !== null && s.act !== filter.act) return false
    if (filter.type === 'checkpoint' && !s.is_checkpoint) return false
    if (filter.type === 'regular' && s.is_checkpoint) return false
    const status = deriveStatus(s, state)
    if (filter.status === 'completed' && status !== 'completed') return false
    if (filter.status === 'pending' && status !== 'pending' && status !== 'needs_confirmation' && status !== 'in_progress') return false
    if (filter.status === 'current' && s.id !== state?.current_step_id) return false
    return true
  })

  const grouped = filtered.reduce<Record<number, GuideStep[]>>((acc, s) => {
    ;(acc[s.act] = acc[s.act] ?? []).push(s)
    return acc
  }, {})

  return (
    <div className="panel">
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '0.6rem' }}>
        <h3 className="panel-title" style={{ margin: 0 }}>Lista kroków</h3>
        <span style={{ fontSize: '0.8rem', color: '#888' }}>
          {state?.confirmed_step_ids.length ?? 0}/{steps.length} zaliczone
        </span>
      </div>

      {/* Filters */}
      <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap', marginBottom: '0.75rem' }}>
        {/* Act filter */}
        <select
          className="filter-select"
          value={filter.act ?? ''}
          onChange={(e) =>
            onFilterChange({ act: e.target.value === '' ? null : Number(e.target.value) })
          }
        >
          <option value="">Wszystkie akty</option>
          {acts.map((a) => (
            <option key={a} value={a}>
              Akt {a}
            </option>
          ))}
        </select>

        {/* Status filter */}
        <select
          className="filter-select"
          value={filter.status}
          onChange={(e) => onFilterChange({ status: e.target.value as StepFilter['status'] })}
        >
          <option value="all">Wszystkie statusy</option>
          <option value="current">Bieżący</option>
          <option value="pending">Oczekujące</option>
          <option value="completed">Zaliczone</option>
        </select>

        {/* Type filter */}
        <select
          className="filter-select"
          value={filter.type}
          onChange={(e) => onFilterChange({ type: e.target.value as StepFilter['type'] })}
        >
          <option value="all">Wszystkie typy</option>
          <option value="checkpoint">Kamienie milowe</option>
          <option value="regular">Zwykłe</option>
        </select>
      </div>

      {filtered.length === 0 && (
        <p className="muted">Brak kroków spełniających filtry.</p>
      )}

      {Object.entries(grouped).map(([actStr, actSteps]) => (
        <section key={actStr} style={{ marginBottom: '1rem' }}>
          <div
            style={{
              fontSize: '0.8rem',
              fontWeight: 600,
              color: '#ffd166',
              textTransform: 'uppercase',
              letterSpacing: '0.06em',
              marginBottom: '0.35rem',
              paddingBottom: '0.2rem',
              borderBottom: '1px solid #2a2a2a',
            }}
          >
            Akt {actStr}
          </div>

          <ol style={{ listStyle: 'none', padding: 0, margin: 0 }}>
            {actSteps.map((step) => {
              const status = deriveStatus(step, state)
              const isCurrent = step.id === state?.current_step_id
              const isExpanded = expandedId === step.id

              return (
                <li
                  key={step.id}
                  style={{
                    marginBottom: '0.25rem',
                    background: isCurrent
                      ? 'rgba(255,209,102,0.07)'
                      : 'rgba(255,255,255,0.02)',
                    borderRadius: 4,
                    border: isCurrent ? '1px solid rgba(255,209,102,0.25)' : '1px solid transparent',
                  }}
                >
                  {/* Step header row */}
                  <div
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: '0.5rem',
                      padding: '0.35rem 0.5rem',
                      cursor: 'pointer',
                    }}
                    onClick={() => setExpandedId(isExpanded ? null : step.id)}
                  >
                    <span
                      style={{ fontSize: '0.75rem', color: '#666', width: 28, flexShrink: 0 }}
                    >
                      {step.step_number}
                    </span>

                    <StatusBadge status={status} small />

                    <span
                      style={{
                        flex: 1,
                        fontSize: '0.875rem',
                        color: status === 'completed' ? '#6ee7b7' : isCurrent ? '#ffd166' : '#e0e0e0',
                        fontWeight: isCurrent ? 600 : 400,
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                        whiteSpace: 'nowrap',
                      }}
                    >
                      {step.title || `Krok ${step.step_number}`}
                    </span>

                    {step.is_checkpoint && (
                      <span
                        style={{
                          fontSize: '0.68rem',
                          color: '#ff6b35',
                          flexShrink: 0,
                        }}
                      >
                        ★
                      </span>
                    )}

                    {step.gem_requirements && step.gem_requirements.length > 0 && (
                      <span style={{ fontSize: '0.7rem', color: '#a8dadc', flexShrink: 0 }}>
                        💎{step.gem_requirements.length}
                      </span>
                    )}

                    <span style={{ fontSize: '0.7rem', color: '#555' }}>
                      {isExpanded ? '▲' : '▼'}
                    </span>
                  </div>

                  {/* Expanded detail */}
                  {isExpanded && (
                    <div
                      style={{
                        padding: '0 0.75rem 0.75rem 2.5rem',
                        borderTop: '1px solid #2a2a2a',
                      }}
                    >
                      <div
                        style={{ fontSize: '0.85rem', color: '#ccc', margin: '0.5rem 0' }}
                        dangerouslySetInnerHTML={{ __html: step.description }}
                      />

                      {step.gem_requirements && step.gem_requirements.length > 0 && (
                        <div style={{ marginTop: '0.5rem' }}>
                          <div style={{ fontSize: '0.75rem', color: '#a8dadc', marginBottom: '0.25rem' }}>
                            💎 Wymagane gemy:
                          </div>
                          <ul style={{ margin: 0, paddingLeft: '1rem' }}>
                            {step.gem_requirements.map((g) => (
                              <li key={g.id} style={{ fontSize: '0.8rem', color: '#ccc' }}>
                                <strong>{g.gem_name}</strong>
                                {g.color && (
                                  <span style={{ marginLeft: '0.3rem', color: '#888' }}>
                                    [{g.color}]
                                  </span>
                                )}
                                {g.note && (
                                  <span style={{ marginLeft: '0.3rem', color: '#999' }}>
                                    — {g.note}
                                  </span>
                                )}
                              </li>
                            ))}
                          </ul>
                        </div>
                      )}

                      {isActive && (
                        <div style={{ display: 'flex', gap: '0.4rem', marginTop: '0.6rem' }}>
                          {isCurrent && (
                            <button
                              className="btn-primary btn-sm"
                              onClick={() => onConfirm(step.id)}
                            >
                              ✓ Potwierdź
                            </button>
                          )}
                          {isCurrent && (
                            <button
                              className="btn-sm"
                              onClick={() => onSkip(step.id)}
                            >
                              Pomiń
                            </button>
                          )}
                          {status === 'completed' && (
                            <button
                              className="btn-sm"
                              onClick={() => onUndo(step.id)}
                            >
                              ↩ Cofnij
                            </button>
                          )}
                        </div>
                      )}
                    </div>
                  )}
                </li>
              )
            })}
          </ol>
        </section>
      ))}
    </div>
  )
}
