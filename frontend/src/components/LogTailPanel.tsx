import { useEffect, useRef, useState } from 'react'
import { subscribeToLogTail } from '../api/client'

const MAX_LINES = 200

const statusColor: Record<string, string> = {
  active: '#6ee7b7',
  waiting_for_file: '#ffd166',
  waiting_for_new_lines: '#ffd166',
  game_not_running: '#888',
  parser_error: '#ff6b35',
  disabled: '#555',
  error: '#ff6b35',
}

interface Props {
  isActive: boolean
}

export function LogTailPanel({ isActive }: Props) {
  const [lines, setLines] = useState<string[]>([])
  const [status, setStatus] = useState<string>('disabled')
  const scrollRef = useRef<HTMLDivElement>(null)
  const autoScroll = useRef(true)

  useEffect(() => {
    const es = subscribeToLogTail()

    es.addEventListener('status', (e) => {
      try {
        const d = JSON.parse((e as MessageEvent).data)
        setStatus(d.status ?? 'unknown')
      } catch {
        // ignore malformed events
      }
    })

    es.addEventListener('log_line', (e) => {
      try {
        const d = JSON.parse((e as MessageEvent).data)
        const line: string = d.line ?? (e as MessageEvent).data
        setLines((prev) => {
          const next = [...prev, line]
          return next.length > MAX_LINES ? next.slice(-MAX_LINES) : next
        })
      } catch {
        // ignore malformed events
      }
    })

    es.onerror = () => setStatus('error')

    return () => es.close()
  }, [])

  // Auto-scroll when new lines arrive, unless the user has scrolled up.
  useEffect(() => {
    const el = scrollRef.current
    if (!el || !autoScroll.current) return
    el.scrollTop = el.scrollHeight
  }, [lines])

  const handleScroll = () => {
    const el = scrollRef.current
    if (!el) return
    const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 40
    autoScroll.current = atBottom
  }

  const dotColor = statusColor[status] ?? '#888'

  return (
    <section className="panel logtail-panel">
      <h3 className="panel-title">
        Client.txt
        <span
          style={{
            marginLeft: '0.6rem',
            color: dotColor,
            fontWeight: 400,
            textTransform: 'none',
            letterSpacing: 0,
          }}
        >
          ● {status.replace(/_/g, ' ')}
        </span>
        {!isActive && (
          <span style={{ marginLeft: '0.5rem', color: '#555', fontWeight: 400, textTransform: 'none', letterSpacing: 0 }}>
            · run nieaktywny
          </span>
        )}
      </h3>

      {lines.length === 0 ? (
        <p className="muted">Oczekiwanie na dane z pliku logu…</p>
      ) : (
        <div
          ref={scrollRef}
          className="logtail-scroll"
          onScroll={handleScroll}
        >
          {lines.map((line, i) => (
            <div key={i} className="logtail-line">
              {line}
            </div>
          ))}
        </div>
      )}
    </section>
  )
}
