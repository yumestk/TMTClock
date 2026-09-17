import type { Today } from './types'
import { fmtDuration, fmtHM, formatElapsed } from './format'

interface Props {
  today: Today | null
  runningElapsed: number
  onRefresh: () => void
}

export default function DayView({ today, runningElapsed, onRefresh }: Props) {
  const sessions = today?.sessions ?? []

  // Total counts every session that started today: finished ones use their
  // recorded span; the running one uses the live elapsed value.
  const finishedSec = sessions
    .filter((s) => s.end_at !== null)
    .reduce((sum, s) => sum + (Date.parse(s.end_at!) - Date.parse(s.start_at)) / 1000, 0)
  const runningInList = sessions.some((s) => s.end_at === null)
  const totalSec = finishedSec + (runningInList ? runningElapsed : 0)

  const maxSec = Math.max(
    runningInList ? runningElapsed : 0,
    ...sessions.filter((s) => s.end_at !== null).map((s) => (Date.parse(s.end_at!) - Date.parse(s.start_at)) / 1000),
    0,
  )

  return (
    <section className="card day">
      <h2>今日</h2>
      {sessions.length === 0 ? (
        <p className="muted">今天还没有记录。</p>
      ) : (
        <ul className="day-list">
          {sessions.map((s) => (
            <DayRow
              key={s.id}
              session={s}
              liveElapsed={s.end_at === null ? runningElapsed : null}
              maxSec={maxSec}
            />
          ))}
        </ul>
      )}
      <footer className="day-total">
        <span>今日合计</span>
        <strong>{sessions.length > 0 ? fmtDuration(totalSec) : '0 分钟'}</strong>
      </footer>
      <button className="day-refresh" onClick={onRefresh}>
        刷新
      </button>
    </section>
  )
}

function DayRow({ session, liveElapsed, maxSec }: { session: Today['sessions'][number]; liveElapsed: number | null; maxSec: number }) {
  const sec = liveElapsed ?? (Date.parse(session.end_at!) - Date.parse(session.start_at)) / 1000
  const expectedSec = session.expected_minutes * 60
  // 2% floor keeps short sessions visible on the proportional bar.
  const width = maxSec > 0 ? Math.max(2, (sec / maxSec) * 100) : 0
  const over = expectedSec > 0 && sec >= expectedSec

  return (
    <li className={`day-row${session.end_at === null ? ' running' : ''}`}>
      <div className="day-meta">
        <span className="day-when">
          {fmtHM(session.start_at)}
          {session.end_at === null ? ' – 进行中' : ` – ${fmtHM(session.end_at)}`}
        </span>
        <span className="day-what">
          {session.activity} / {session.project}
        </span>
        {session.note && <span className="day-note">{session.note}</span>}
        <span className={`day-dur${over ? ' over' : ''}`}>
          {session.end_at === null ? formatElapsed(liveElapsed ?? 0) : fmtDuration(sec)}
        </span>
      </div>
      <div className="day-bar">
        <div
          className={`day-fill${session.end_at === null ? ' running' : ''}${over ? ' over' : ''}`}
          style={{ width: `${width}%` }}
        />
        {expectedSec > 0 && maxSec > 0 && (
          <span
            className="day-tick"
            style={{ left: `${Math.min(100, (expectedSec / maxSec) * 100)}%` }}
            title={`预期 ${session.expected_minutes} 分钟`}
          />
        )}
      </div>
    </li>
  )
}
