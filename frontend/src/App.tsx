import { useCallback, useEffect, useRef, useState } from 'react'
import type { Activity, Session, Today } from './types'
import * as api from './api'
import ManagePanel from './ManagePanel'
import TimerCard from './TimerCard'
import DayView from './DayView'
import { useElapsed } from './useElapsed'
import { chime, notifyOver } from './reminder'

function GearIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" aria-hidden>
      <circle cx="12" cy="12" r="3.2" />
      <path
        d="M12 2.8v2.4M12 18.8v2.4M4.6 6.9l1.7 1M17.7 16.1l1.7 1M4.6 17.1l1.7-1M17.7 7.9l1.7-1M21.2 12h-2.4M5.2 12H2.8"
        strokeLinecap="round"
      />
    </svg>
  )
}

export default function App() {
  const [activities, setActivities] = useState<Activity[]>([])
  const [running, setRunning] = useState<Session | null>(null)
  const [today, setToday] = useState<Today | null>(null)
  const [loadError, setLoadError] = useState('')
  const [manageOpen, setManageOpen] = useState(false)

  const refetchActivities = useCallback(() => {
    api.listActivities().then(setActivities).catch((e) => setLoadError(String(e)))
  }, [])

  const refetchRunning = useCallback(() => {
    api.getRunning().then(setRunning).catch((e) => setLoadError(String(e)))
  }, [])

  const refetchToday = useCallback(() => {
    api.getToday().then(setToday).catch((e) => setLoadError(String(e)))
  }, [])

  useEffect(() => {
    refetchActivities()
    refetchRunning()
    refetchToday()
  }, [refetchActivities, refetchRunning, refetchToday])

  // One shared tick drives both the timer card and the live day-view row.
  const runningElapsed = useElapsed(running?.start_at ?? null)

  // Reminder fires once when the running session crosses its expected
  // duration. A session already over the mark when first observed (page
  // refresh) goes straight to the armed state — visual over state without
  // replaying the chime.
  const armRef = useRef<{ id: number; armed: boolean } | null>(null)
  useEffect(() => {
    if (!running || running.expected_minutes <= 0) {
      armRef.current = null
      return
    }
    const threshold = running.expected_minutes * 60
    const over = runningElapsed >= threshold
    let st = armRef.current
    if (!st || st.id !== running.id) {
      st = { id: running.id, armed: over }
      armRef.current = st
    }
    if (over && !st.armed) {
      chime()
      notifyOver(running.activity, running.project)
      st.armed = true
    }
  }, [running, runningElapsed])

  return (
    <div className="app">
      <header className="app-header">
        <h1>
          ToMaTo<span>Clock</span>
        </h1>
        <button className="gear" onClick={() => setManageOpen(true)} aria-label="管理">
          <GearIcon />
        </button>
      </header>
      {loadError && <p className="error">{loadError}（后端在跑吗？make dev-go）</p>}
      <main>
        <TimerCard
          activities={activities}
          running={running}
          onStarted={(s) => {
            setRunning(s)
            refetchToday()
          }}
          onStopped={() => {
            refetchRunning()
            refetchToday()
          }}
        />
        <DayView today={today} runningElapsed={runningElapsed} />
      </main>
      <ManagePanel
        activities={activities}
        open={manageOpen}
        onClose={() => setManageOpen(false)}
        onChanged={refetchActivities}
      />
    </div>
  )
}
