// Formats seconds as H:MM:SS (hours unbounded, no leading zero).
export function formatElapsed(seconds: number): string {
  const total = Math.max(0, Math.floor(seconds))
  const h = Math.floor(total / 3600)
  const m = Math.floor((total % 3600) / 60)
  const s = total % 60
  return `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

// Local HH:MM for a timestamp string.
export function fmtHM(iso: string): string {
  const d = new Date(iso)
  return `${pad(d.getHours())}:${pad(d.getMinutes())}`
}

// Compact human duration: seconds under a minute, minutes under an hour,
// otherwise hours + minutes.
export function fmtDuration(sec: number): string {
  const total = Math.max(0, Math.floor(sec))
  if (total < 60) return `${total} 秒`
  const h = Math.floor(total / 3600)
  const m = Math.floor((total % 3600) / 60)
  if (h === 0) return `${m} 分钟`
  return m > 0 ? `${h} 小时 ${m} 分` : `${h} 小时`
}

function pad(n: number): string {
  return String(n).padStart(2, '0')
}
