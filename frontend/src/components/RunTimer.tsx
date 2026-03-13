import { useEffect, useRef, useState } from 'react'

interface Props {
  startedAt: string
  isActive: boolean
  serverElapsedMs: number
}

/** Displays a live timer counting up from the run's start time. */
export function RunTimer({ startedAt, isActive, serverElapsedMs }: Props) {
  const [elapsedMs, setElapsedMs] = useState(serverElapsedMs)
  const ref = useRef<ReturnType<typeof setInterval> | null>(null)

  useEffect(() => {
    if (!isActive) {
      setElapsedMs(serverElapsedMs)
      return
    }
    const startTs = new Date(startedAt).getTime()
    const tick = () => setElapsedMs(Date.now() - startTs)
    tick()
    ref.current = setInterval(tick, 1000)
    return () => {
      if (ref.current) clearInterval(ref.current)
    }
  }, [startedAt, isActive, serverElapsedMs])

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
