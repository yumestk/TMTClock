import { useState } from 'react'
import type { Activity, Session } from './types'
import * as api from './api'
import { useElapsed } from './useElapsed'
import { formatElapsed } from './format'
import { primeReminder } from './reminder'

interface Props {
  activities: Activity[]
  running: Session | null
  onStarted: (session: Session) => void
  onStopped: () => void
}

const R = 108
const CIRC = 2 * Math.PI * R

export default function TimerCard({ activities, running, onStarted, onStopped }: Props) {
  if (running) return <RunningCard running={running} onStopped={onStopped} />
  return <IdleCard activities={activities} onStarted={onStarted} />
}

function Ring({
  progress,
  idle,
  children,
}: {
  progress: number
  idle: boolean
  children: React.ReactNode
}) {
  return (
    <div className={`ring${idle ? ' idle' : ''}`}>
      <svg viewBox="0 0 236 236">
        <circle className="ring-track" cx="118" cy="118" r={R} />
        {!idle && (
          <circle
            className="ring-fill"
            cx="118"
            cy="118"
            r={R}
            strokeDasharray={CIRC}
            strokeDashoffset={CIRC * (1 - Math.min(1, progress))}
          />
        )}
      </svg>
      {children}
    </div>
  )
}

function parseMinutes(text: string): number {
  const n = Number(text)
  return Number.isFinite(n) && n > 0 ? Math.floor(n) : 0
}

function IdleCard({ activities, onStarted }: { activities: Activity[]; onStarted: (s: Session) => void }) {
  const [projectId, setProjectId] = useState(0)
  // null = untouched by the user; the field then shows the project default.
  const [expected, setExpected] = useState<string | null>(null)
  const [note, setNote] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  const projects = activities.flatMap((a) => a.projects)
  const selected = projects.find((p) => p.id === projectId) ?? projects[0]
  const expectedValue = expected ?? (selected ? String(selected.expected_minutes) : '')

  const onStart = async () => {
    if (!selected) return
    // Inside the click gesture: unlock audio (autoplay policy) and ask for
    // notification permission before any reminder could be due.
    primeReminder(parseMinutes(expectedValue) > 0)
    setBusy(true)
    setError('')
    try {
      const s = await api.startSession(selected.id, note, parseMinutes(expectedValue))
      onStarted(s)
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e))
    } finally {
      setBusy(false)
    }
  }

  return (
    <section className="card timer idle">
      <h2 className="timer-title">
        计时<span className="jp-label">けいそく</span>
      </h2>
      <Ring progress={0} idle>
        <span className="ring-idle-label">待機中…</span>
      </Ring>
      {projects.length === 0 ? (
        <p className="muted">还没有项目。点右上角齿轮，先建一个分类和项目。</p>
      ) : (
        <div className="timer-form">
          <label>
            项目
            <select
              value={selected.id}
              onChange={(e) => {
                setProjectId(Number(e.target.value))
                setExpected(null)
              }}
            >
              {activities.map((a) =>
                a.projects.length > 0 ? (
                  <optgroup key={a.id} label={a.name}>
                    {a.projects.map((p) => (
                      <option key={p.id} value={p.id}>
                        {p.name}
                        {p.expected_minutes > 0 ? ` · ${p.expected_minutes} 分钟` : ''}
                      </option>
                    ))}
                  </optgroup>
                ) : null,
              )}
            </select>
          </label>
          <label>
            预期分钟
            <input
              className="narrow"
              inputMode="numeric"
              value={expectedValue}
              onChange={(e) => setExpected(e.target.value)}
              placeholder="0 = 不限"
            />
          </label>
          <label>
            备注
            <input value={note} onChange={(e) => setNote(e.target.value)} placeholder="这次做什么" />
          </label>
          {error && <p className="error">{error}</p>}
          <button className="primary" onClick={onStart} disabled={busy}>
            开始
          </button>
        </div>
      )}
    </section>
  )
}

// Zen Maru Gothic digits are not tabular (font-variant-numeric has no
// effect), so each char gets a fixed-width slot to stop the readout from
// wobbling every second.
function Elapsed({ text }: { text: string }) {
  return (
    <div className="elapsed">
      {text.split('').map((ch, i) => (
        <span key={i} className={ch === ':' ? 'c' : 'd'}>
          {ch}
        </span>
      ))}
    </div>
  )
}

function RunningCard({ running, onStopped }: { running: Session; onStopped: () => void }) {
  const elapsed = useElapsed(running.start_at)
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  const onStop = async () => {
    setBusy(true)
    setError('')
    try {
      await api.stopSession()
      onStopped()
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e))
    } finally {
      setBusy(false)
    }
  }

  const expectedSec = running.expected_minutes * 60
  const overExpected = expectedSec > 0 && elapsed >= expectedSec
  const progress = expectedSec > 0 ? elapsed / expectedSec : 0

  return (
    <section className={`card timer running${overExpected ? ' over' : ''}`}>
      <h2 className="timer-title">
        {running.activity} / {running.project}
      </h2>
      {running.note && <p className="muted timer-note">{running.note}</p>}
      <Ring progress={progress} idle={false}>
        <Elapsed text={formatElapsed(elapsed)} />
      </Ring>
      {expectedSec > 0 ? (
        <p className={`timer-status${overExpected ? ' over' : ''}`}>
          {overExpected
            ? `已超预期 ${Math.floor((elapsed - expectedSec) / 60)} 分`
            : `预期 ${running.expected_minutes} 分钟 · 剩余 ${Math.ceil((expectedSec - elapsed) / 60)} 分`}
        </p>
      ) : (
        <p className="timer-status">計測中…</p>
      )}
      {error && <p className="error">{error}</p>}
      <button className="primary stop" onClick={onStop} disabled={busy}>
        停止
      </button>
    </section>
  )
}
