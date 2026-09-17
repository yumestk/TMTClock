import { useCallback, useEffect, useState } from 'react'
import type { Activity, Session } from './types'
import * as api from './api'
import ManagePanel from './ManagePanel'

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
        {/* TimerCard and DayView arrive in steps 5 and 6. */}
        {running ? (
          <p className="muted">计时进行中：{running.activity} / {running.project}（计时卡片即将上线）</p>
        ) : (
          <p className="muted">空闲中。先在下方管理面板建一个分类和项目。</p>
        )}
        <ManagePanel activities={activities} onChanged={refetchActivities} />
      </main>
    </div>
  )
}
