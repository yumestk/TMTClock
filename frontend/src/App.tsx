import { useCallback, useEffect, useState } from 'react'
import type { Activity, Session, Today } from './types'
import * as api from './api'
import ManagePanel from './ManagePanel'
import TimerCard from './TimerCard'
import DayView from './DayView'
import { useElapsed } from './useElapsed'

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
