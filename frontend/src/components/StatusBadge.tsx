import type { StepProgressStatus } from '../api/types'

const config: Record<StepProgressStatus, { label: string; color: string; bg: string }> = {
  completed: { label: 'Zaliczone', color: '#fff', bg: '#1a6b45' },
  in_progress: { label: 'W toku', color: '#fff', bg: '#1a4a6b' },
  needs_confirmation: { label: 'Wymaga potwierdzenia', color: '#fff', bg: '#8a4a00' },
  skipped: { label: 'Pominięte', color: '#aaa', bg: '#2a2a2a' },
  pending: { label: 'Oczekujące', color: '#888', bg: '#222' },
}

interface Props {
  status: StepProgressStatus
  small?: boolean
}

export function StatusBadge({ status, small }: Props) {
  const { label, color, bg } = config[status] ?? config.pending
  return (
    <span
      style={{
        display: 'inline-block',
        padding: small ? '0.1rem 0.4rem' : '0.15rem 0.55rem',
        borderRadius: 3,
        fontSize: small ? '0.7rem' : '0.75rem',
        fontWeight: 500,
        letterSpacing: '0.02em',
        background: bg,
        color,
        flexShrink: 0,
      }}
    >
      {label}
    </span>
  )
}
