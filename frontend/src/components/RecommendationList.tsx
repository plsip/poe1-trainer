import type { Recommendation } from '../api/types'

const priorityColor: Record<string, string> = {
  high: '#ff6b35',
  medium: '#ffd166',
  low: '#a8dadc',
}

interface Props {
  recommendations: Recommendation[]
}

export function RecommendationList({ recommendations }: Props) {
  if (recommendations.length === 0) {
    return <p style={{ color: '#888' }}>Brak rekomendacji.</p>
  }
  return (
    <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
      {recommendations.map((r) => (
        <li
          key={r.id}
          style={{
            borderLeft: `4px solid ${priorityColor[r.priority] ?? '#ccc'}`,
            padding: '0.6rem 1rem',
            marginBottom: '0.5rem',
            background: 'rgba(255,255,255,0.04)',
            borderRadius: '0 4px 4px 0',
          }}
        >
          <strong style={{ display: 'block', marginBottom: '0.2rem' }}>{r.text}</strong>
          <span style={{ fontSize: '0.85rem', color: '#b0b0b0' }}>{r.reason}</span>
        </li>
      ))}
    </ul>
  )
}
