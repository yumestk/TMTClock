import { useCallback, useEffect, useRef, useState } from 'react'
import type { Activity, Session, Today } from './types'
import * as api from './api'
import ManagePanel from './ManagePanel'
import TimerCard from './TimerCard'
import DayView from './DayView'
import { useElapsed } from './useElapsed'
import { chime, notifyOver } from './reminder'

export default function App() {
  const [activities, setActivities] = useState<Activity[]>([])
  const [running, setRunning] = useState<Session | null>(null)
  const [today, setToday] = useState<Today | null>(null)
  const [loadError, setLoadError] = useState('')

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
      <header>
        <h1>ToMaToClock</h1>
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
        <DayView
          today={today}
          running={running}
          runningElapsed={runningElapsed}
          onRefresh={refetchToday}
        />
        <ManagePanel activities={activities} onChanged={refetchActivities} />
      </main>
    </div>
  )
}
