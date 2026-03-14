import { useEffect, useRef, useState } from 'react'

interface Props {
  /** Whether the timer is actively counting (run active AND not paused). */
  isActive: boolean
  /** Server-authoritative elapsed milliseconds at the last state fetch. */
  serverElapsedMs: number
}

/** Displays a live timer counting up from serverElapsedMs when active, frozen when paused. */
export function RunTimer({ isActive, serverElapsedMs }: Props) {
  const [elapsedMs, setElapsedMs] = useState(serverElapsedMs)
  const ref = useRef<ReturnType<typeof setInterval> | null>(null)

  useEffect(() => {
    if (!isActive) {
      setElapsedMs(serverElapsedMs)
      return
    }
    // Anchor live counting to the server value at the time this effect runs.
    const baseMs = serverElapsedMs
    const baseTs = Date.now()
    const tick = () => setElapsedMs(baseMs + (Date.now() - baseTs))
    tick()
    ref.current = setInterval(tick, 1000)
    return () => {
      if (ref.current) clearInterval(ref.current)
    }
  }, [isActive, serverElapsedMs])

  const totalSec = Math.floor(elapsedMs / 1000)
  const h = Math.floor(totalSec / 3600)
  const m = Math.floor((totalSec % 3600) / 60)
  const s = totalSec % 60
  const label = h > 0
    ? `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
    : `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`

  return (
    <span style={{ fontFamily: 'monospace', fontSize: '1.4rem' }}>
      ⏱ {label}
    </span>
  )
}
