import { useCallback, useEffect, useState } from 'react'
import type { Activity, Session } from './types'
import * as api from './api'
import ManagePanel from './ManagePanel'
import TimerCard from './TimerCard'

export default function App() {
  const [activities, setActivities] = useState<Activity[]>([])
  const [running, setRunning] = useState<Session | null>(null)
  const [loadError, setLoadError] = useState('')

  const refetchActivities = useCallback(() => {
    api.listActivities().then(setActivities).catch((e) => setLoadError(String(e)))
  }, [])

  const refetchRunning = useCallback(() => {
    api.getRunning().then(setRunning).catch((e) => setLoadError(String(e)))
  }, [])

  useEffect(() => {
    refetchActivities()
    refetchRunning()
  }, [refetchActivities, refetchRunning])

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
          onStarted={setRunning}
          onStopped={refetchRunning}
        />
        {/* DayView arrives in step 6. */}
        <ManagePanel activities={activities} onChanged={refetchActivities} />
      </main>
    </div>
  )
}
