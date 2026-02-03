import { useEffect, useMemo, useState } from 'react'
import { downloadTaskReport, listTaskDeployments, type TaskDeploymentDetail } from '../api/client'
import { useDashboardData } from '../state/use-dashboard-data'

interface DeploymentsProps {
  isDemoMode: boolean
  statusFilter: 'success' | 'failed' | null
  onCreateDeployment: () => void
}

export function Deployments({ isDemoMode, statusFilter, onCreateDeployment }: DeploymentsProps) {
  const { tasks, isLoading, errorMessage } = useDashboardData({ isDemoMode })
  const [selectedTaskId, setSelectedTaskId] = useState<string>('')
  const [localStatusFilter, setLocalStatusFilter] = useState<'all' | 'success' | 'failed'>('all')
  const [deployments, setDeployments] = useState<TaskDeploymentDetail[]>([])
  const [deploymentsError, setDeploymentsError] = useState<string | null>(null)
  const [isDeploymentsLoading, setIsDeploymentsLoading] = useState(false)

  const effectiveFilter = statusFilter ?? (localStatusFilter === 'all' ? null : localStatusFilter)

  const filteredTasks = useMemo(() => {
    if (!effectiveFilter) return tasks
    return tasks.filter((task) => task.status === effectiveFilter)
  }, [tasks, effectiveFilter])

  useEffect(() => {
    if (!selectedTaskId && filteredTasks.length > 0) {
      setSelectedTaskId(filteredTasks[0].id)
    }
  }, [filteredTasks, selectedTaskId])

  useEffect(() => {
    if (!selectedTaskId) return

    if (isDemoMode) {
      const filtered = applyStatusFilter(demoTaskDeployments, effectiveFilter)
      setDeployments(filtered)
      setDeploymentsError(null)
      return
    }

    let isMounted = true
    setIsDeploymentsLoading(true)
    setDeploymentsError(null)

    listTaskDeployments(selectedTaskId)
      .then((response) => {
        if (!isMounted) return
        setDeployments(applyStatusFilter(response, effectiveFilter))
      })
      .catch((error) => {
        if (!isMounted) return
        setDeploymentsError(error instanceof Error ? error.message : 'Failed to load deployments')
      })
      .finally(() => {
        if (!isMounted) return
        setIsDeploymentsLoading(false)
      })

    return () => {
      isMounted = false
    }
  }, [selectedTaskId, isDemoMode, effectiveFilter])

  const handleExport = async (format: 'csv' | 'pdf') => {
    if (!selectedTaskId) return
    const blob = await downloadTaskReport(selectedTaskId, format)
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `task-${selectedTaskId}-deployments.${format}`
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
  }

  if (isLoading) {
    return <div className="rounded-2xl border border-slate-800 bg-slate-900/60 p-6">Loading tasks...</div>
  }

  if (errorMessage) {
    return (
      <div className="rounded-2xl border border-slate-800 bg-slate-900/60 p-6 text-rose-300">
        {errorMessage}
      </div>
    )
  }

  return (
    <section className="rounded-2xl border border-slate-800 bg-slate-900/60 p-6 shadow-xl">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <p className="text-sm uppercase tracking-[0.25em] text-slate-400">Deployments</p>
          <h2 className="text-2xl font-semibold text-white">Deployment details</h2>
        </div>
        <div className="flex items-center gap-3">
          <button
            type="button"
            onClick={onCreateDeployment}
            className="rounded-lg border border-indigo-400 bg-indigo-500/20 px-3 py-2 text-xs text-indigo-200"
          >
            Create Deployment Task
          </button>
        </div>
      </div>

      <div className="mt-6 flex flex-wrap items-center gap-3">
        <select
          className="rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-xs text-slate-200"
          value={localStatusFilter}
          onChange={(event) => setLocalStatusFilter(event.target.value as 'all' | 'success' | 'failed')}
        >
          <option value="all">All statuses</option>
          <option value="success">Success only</option>
          <option value="failed">Failed only</option>
        </select>
        <button
          type="button"
          onClick={() => handleExport('csv')}
          disabled={!selectedTaskId}
          className="rounded-lg border border-slate-700 px-3 py-2 text-xs text-slate-200 disabled:cursor-not-allowed disabled:opacity-50"
        >
          Export CSV
        </button>
        <button
          type="button"
          onClick={() => handleExport('pdf')}
          disabled={!selectedTaskId}
          className="rounded-lg border border-slate-700 px-3 py-2 text-xs text-slate-200 disabled:cursor-not-allowed disabled:opacity-50"
        >
          Export PDF
        </button>
      </div>

      <div className="mt-4 space-y-2 text-xs text-slate-400">
        {isDeploymentsLoading && <p>Loading deployments...</p>}
        {deploymentsError && <p className="text-rose-300">{deploymentsError}</p>}
        {!isDeploymentsLoading && !deploymentsError && deployments.length === 0 && (
          <p>No deployment results yet.</p>
        )}
      </div>

      <div className="mt-4 grid gap-3 md:grid-cols-2">
        {filteredTasks.map((task) => (
          <button
            key={task.id}
            type="button"
            onClick={() => setSelectedTaskId(task.id)}
            className={`rounded-lg border px-4 py-3 text-left text-sm ${
              selectedTaskId === task.id
                ? 'border-indigo-400 bg-indigo-500/20 text-indigo-200'
                : 'border-slate-800 bg-slate-900/80 text-slate-300'
            }`}
          >
            <div className="flex items-center justify-between">
              <span className="font-semibold">{task.name}</span>
              <span className="text-xs text-slate-400">{task.status}</span>
            </div>
            <p className="mt-1 text-xs text-slate-400">Targets: {task.targetCount}</p>
          </button>
        ))}
        {filteredTasks.length === 0 && (
          <div className="rounded-lg border border-slate-800 bg-slate-900/80 p-4 text-xs text-slate-400">
            No tasks match this filter.
          </div>
        )}
      </div>

      {deployments.length > 0 && (
        <div className="mt-3 overflow-hidden rounded-lg border border-slate-800">
          <table className="w-full text-left text-xs text-slate-300">
            <thead className="bg-slate-900/80 text-slate-400">
              <tr>
                <th className="px-3 py-2">Target</th>
                <th className="px-3 py-2">OS</th>
                <th className="px-3 py-2">Status</th>
                <th className="px-3 py-2">Auth</th>
                <th className="px-3 py-2">Error</th>
                <th className="px-3 py-2">Finished</th>
              </tr>
            </thead>
            <tbody>
              {deployments.map((deployment) => (
                <tr key={deployment.id} className="border-t border-slate-800">
                  <td className="px-3 py-2">{deployment.targetLabel}</td>
                  <td className="px-3 py-2">{deployment.targetOS}</td>
                  <td className="px-3 py-2">{deployment.status}</td>
                  <td className="px-3 py-2">{deployment.authMethod}</td>
                  <td className="px-3 py-2 text-rose-300">{deployment.errorCode || '-'}</td>
                  <td className="px-3 py-2">{deployment.finishedAt}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  )
}

const demoTaskDeployments: TaskDeploymentDetail[] = [
  {
    id: 'demo-deploy-1',
    taskRunId: 'demo-run',
    targetId: 'demo-target-1',
    targetLabel: 'core-db-01',
    targetOS: 'linux',
    status: 'success',
    authMethod: 'ssh_key',
    errorCode: '',
    errorMessage: '',
    remediation: '',
    finishedAt: '2026-02-02T10:45:00Z'
  },
  {
    id: 'demo-deploy-2',
    taskRunId: 'demo-run',
    targetId: 'demo-target-2',
    targetLabel: 'win-edge-04',
    targetOS: 'windows',
    status: 'failed',
    authMethod: 'winrm_https_cert',
    errorCode: 'auth_denied',
    errorMessage: 'Authentication denied',
    remediation: 'Verify credentials and retry with certificate auth',
    finishedAt: '2026-02-02T10:46:00Z'
  }
]

function applyStatusFilter(
  deployments: TaskDeploymentDetail[],
  filter: 'success' | 'failed' | null
) {
  if (!filter) return deployments
  return deployments.filter((row) => row.status === filter)
}
