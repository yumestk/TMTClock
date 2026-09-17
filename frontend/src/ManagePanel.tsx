import { useState } from 'react'
import type { Activity } from './types'
import * as api from './api'

interface Props {
  activities: Activity[]
  onChanged: () => void
}

export default function ManagePanel({ activities, onChanged }: Props) {
  const [open, setOpen] = useState(false)
  const projectCount = activities.reduce((n, a) => n + a.projects.length, 0)

  if (!open) {
    return (
      <section className="manage">
        <button className="manage-toggle" onClick={() => setOpen(true)}>
          管理 · {activities.length} 分类 · {projectCount} 项目
        </button>
      </section>
    )
  }

  return (
    <section className="manage open">
      <button className="manage-toggle" onClick={() => setOpen(false)}>
        收起管理
      </button>
      <AddActivityForm onAdded={onChanged} />
      {activities.map((a) => (
        <ActivityRow key={a.id} activity={a} onChanged={onChanged} />
      ))}
    </section>
  )
}

function AddActivityForm({ onAdded }: { onAdded: () => void }) {
  const [name, setName] = useState('')
  const [error, setError] = useState('')

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) return
    try {
      await api.createActivity(name.trim())
      setName('')
      setError('')
      onAdded()
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    }
  }

  return (
    <form className="manage-add" onSubmit={submit}>
      <input
        value={name}
        onChange={(e) => setName(e.target.value)}
        placeholder="新分类名（如 学习 / 工作）"
      />
      <button type="submit">添加分类</button>
      {error && <p className="error">{error}</p>}
    </form>
  )
}

function ActivityRow({ activity, onChanged }: { activity: Activity; onChanged: () => void }) {
  const [renaming, setRenaming] = useState(false)
  const [name, setName] = useState(activity.name)
  const [error, setError] = useState('')

  const rename = async () => {
    if (!name.trim() || name.trim() === activity.name) {
      setRenaming(false)
      setName(activity.name)
      return
    }
    try {
      await api.renameActivity(activity.id, name.trim())
      setError('')
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    }
    setRenaming(false)
    onChanged()
  }

  const remove = async () => {
    if (!confirm(`删除分类「${activity.name}」及其下所有项目？`)) return
    try {
      await api.deleteActivity(activity.id)
      setError('')
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    }
    onChanged()
  }

  return (
    <div className="manage-activity">
      <div className="manage-activity-head">
        {renaming ? (
          <input
            autoFocus
            value={name}
            onChange={(e) => setName(e.target.value)}
            onBlur={rename}
            onKeyDown={(e) => e.key === 'Enter' && rename()}
          />
        ) : (
          <strong>{activity.name}</strong>
        )}
        <span className="manage-actions">
          <button onClick={() => setRenaming(!renaming)}>{renaming ? '取消' : '改名'}</button>
          <button onClick={remove}>删除</button>
        </span>
      </div>
      {error && <p className="error">{error}</p>}
      {activity.projects.map((p) => (
        <ProjectRow key={p.id} project={p} onChanged={onChanged} />
      ))}
      <AddProjectForm activityId={activity.id} onAdded={onChanged} />
    </div>
  )
}

function ProjectRow({ project, onChanged }: { project: Activity['projects'][number]; onChanged: () => void }) {
  const [editing, setEditing] = useState(false)
  const [name, setName] = useState(project.name)
  const [expected, setExpected] = useState(String(project.expected_minutes))
  const [error, setError] = useState('')

  const save = async () => {
    const expectedNum = Number(expected) || 0
    if (expectedNum < 0) {
      setError('预期分钟不能为负')
      return
    }
    try {
      await api.updateProject(project.id, {
        name: name.trim() || project.name,
        expected_minutes: expectedNum,
      })
      setError('')
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    }
    setEditing(false)
    onChanged()
  }

  const remove = async () => {
    if (!confirm(`删除项目「${project.name}」？`)) return
    try {
      await api.deleteProject(project.id)
      setError('')
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    }
    onChanged()
  }

  if (!editing) {
    return (
      <div className="manage-project">
        <span>{project.name}</span>
        <span className="muted">
          {project.expected_minutes > 0 ? `默认 ${project.expected_minutes} 分钟` : '无默认时长'}
        </span>
        <span className="manage-actions">
          <button onClick={() => setEditing(true)}>编辑</button>
          <button onClick={remove}>删除</button>
        </span>
      </div>
    )
  }

  return (
    <div className="manage-project editing">
      <input value={name} onChange={(e) => setName(e.target.value)} />
      <input
        type="number"
        min={0}
        value={expected}
        onChange={(e) => setExpected(e.target.value)}
        placeholder="默认分钟"
        className="narrow"
      />
      <span className="manage-actions">
        <button onClick={save}>保存</button>
        <button onClick={() => setEditing(false)}>取消</button>
      </span>
      {error && <p className="error">{error}</p>}
    </div>
  )
}

function AddProjectForm({ activityId, onAdded }: { activityId: number; onAdded: () => void }) {
  const [name, setName] = useState('')
  const [expected, setExpected] = useState('')
  const [error, setError] = useState('')

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) return
    try {
      await api.createProject(activityId, name.trim(), Number(expected) || 0)
      setName('')
      setExpected('')
      setError('')
      onAdded()
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    }
  }

  return (
    <form className="manage-add sub" onSubmit={submit}>
      <input value={name} onChange={(e) => setName(e.target.value)} placeholder="新项目名" />
      <input
        type="number"
        min={0}
        value={expected}
        onChange={(e) => setExpected(e.target.value)}
        placeholder="默认分钟（可空）"
        className="narrow"
      />
      <button type="submit">添加项目</button>
      {error && <p className="error">{error}</p>}
    </form>
  )
}
