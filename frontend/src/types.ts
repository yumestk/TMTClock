export interface Project {
  id: number
  activity_id: number
  name: string
  expected_minutes: number
}

export interface Activity {
  id: number
  name: string
  projects: Project[]
}

export interface Session {
  id: number
  activity: string
  project: string
  note: string
  expected_minutes: number
  start_at: string
  end_at: string | null
}

export interface Today {
  date: string
  sessions: Session[]
}
