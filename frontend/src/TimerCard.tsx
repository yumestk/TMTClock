import { useState } from 'react'
import type { Activity, Session } from './types'
import * as api from './api'
import { useElapsed } from './useElapsed'
import { formatElapsed } from './format'

interface Props {
  activities: Activity[]
  running: Session | null
  onStarted: (session: Session) => void
  onStopped: () => void
}

export default function TimerCard({ activities, running, onStarted, onStopped }: Props) {
  if (running) return <RunningCard running={running} onStopped={onStopped} />
  return <IdleCard activities={activities} onStarted={onStarted} />
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
    <section className="card timer">
      <h2>计时</h2>
      {projects.length === 0 ? (
        <p className="muted">还没有项目。先在下方管理面板建一个分类和项目。</p>
      ) : (
        <>
          <label className="field">
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
          <label className="field">
            预期分钟
            <input
              className="narrow"
              inputMode="numeric"
              value={expectedValue}
              onChange={(e) => setExpected(e.target.value)}
              placeholder="0 = 不限"
            />
          </label>
          <label className="field">
            备注
            <input value={note} onChange={(e) => setNote(e.target.value)} placeholder="这次做什么" />
          </label>
          {error && <p className="error">{error}</p>}
          <button className="primary" onClick={onStart} disabled={busy}>
            开始
          </button>
        </>
      )}
    </section>
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

  return (
    <section className={`card timer${overExpected ? ' over' : ''}`}>
      <h2>
        {running.activity} / {running.project}
      </h2>
      {running.note && <p className="muted">{running.note}</p>}
      <div className="elapsed">{formatElapsed(elapsed)}</div>
      {expectedSec > 0 && (
        <p className="muted">
          预期 {running.expected_minutes} 分钟 ·{' '}
          {overExpected ? '已超预期' : `剩余 ${Math.ceil((expectedSec - elapsed) / 60)} 分钟`}
        </p>
      )}
      {error && <p className="error">{error}</p>}
      <button className="primary stop" onClick={onStop} disabled={busy}>
        停止
      </button>
    </section>
  )
}
