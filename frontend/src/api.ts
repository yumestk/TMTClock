import type { Activity, Project, Session, Today } from './types'

class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const res = await fetch(path, {
    method,
    headers: body !== undefined ? { 'Content-Type': 'application/json' } : undefined,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })
  if (!res.ok) {
    let message = `${res.status} ${res.statusText}`
    try {
      const data = await res.json()
      if (data.error) message = data.error
    } catch {
      // keep default message
    }
    throw new ApiError(res.status, message)
  }
  if (res.status === 204) return null as T
  return res.json() as Promise<T>
}

export const listActivities = () => request<Activity[]>('GET', '/api/activities')

export const createActivity = (name: string) =>
  request<Activity>('POST', '/api/activities', { name })

export const renameActivity = (id: number, name: string) =>
  request<Activity>('PATCH', `/api/activities/${id}`, { name })

export const deleteActivity = (id: number) =>
  request<void>('DELETE', `/api/activities/${id}`)

export const createProject = (activityId: number, name: string, expectedMinutes: number) =>
  request<Project>('POST', '/api/projects', {
    activity_id: activityId,
    name,
    expected_minutes: expectedMinutes,
  })

export const updateProject = (id: number, patch: { name?: string; expected_minutes?: number }) =>
  request<Project>('PATCH', `/api/projects/${id}`, patch)

export const deleteProject = (id: number) =>
  request<void>('DELETE', `/api/projects/${id}`)

export const getRunning = () =>
  request<Session | null>('GET', '/api/running')

export const startSession = (projectId: number, note: string, expectedMinutes: number) =>
  request<Session>('POST', '/api/sessions/start', {
    project_id: projectId,
    note,
    expected_minutes: expectedMinutes,
  })

export const stopSession = (note?: string) =>
  request<Session>('POST', '/api/sessions/stop', note !== undefined ? { note } : {})

export const getToday = () => request<Today>('GET', '/api/today')

export { ApiError }
