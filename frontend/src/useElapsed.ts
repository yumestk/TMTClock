import { useEffect, useState } from 'react'

// Seconds elapsed since startAt, recomputed on a 1s tick. Elapsed is always
// derived from the timestamp, never accumulated, so it stays correct after
// refreshes, sleeps, and background-tab throttling. null = nothing running:
// no interval runs and the value stays 0.
export function useElapsed(startAt: string | null): number {
  const [now, setNow] = useState(() => Date.now())
  useEffect(() => {
    if (startAt === null) return
    const refresh = () => setNow(Date.now())
    const id = setInterval(refresh, 1000)
    document.addEventListener('visibilitychange', refresh)
    return () => {
      clearInterval(id)
      document.removeEventListener('visibilitychange', refresh)
    }
  }, [startAt])
  if (startAt === null) return 0
  return Math.max(0, (now - Date.parse(startAt)) / 1000)
}
