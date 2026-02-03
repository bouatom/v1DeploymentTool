import { useState } from 'react'
import { Dashboard } from './components/Dashboard'
import { Deployments } from './components/Deployments'
import { Wizard } from './components/Wizard'

type Page = 'dashboard' | 'deploy' | 'deployments'
type DeploymentFilter = 'success' | 'failed' | null

export function App() {
  const [refreshKey, setRefreshKey] = useState(0)
  const [isDemoMode, setIsDemoMode] = useState(false)
  const [page, setPage] = useState<Page>('dashboard')
  const [deploymentFilter, setDeploymentFilter] = useState<DeploymentFilter>(null)

  const handleTaskCreated = () => {
    setRefreshKey((value) => value + 1)
  }

  const handleViewDeployments = (filter: DeploymentFilter) => {
    setDeploymentFilter(filter)
    setPage('deployments')
  }

  return (
    <div className="min-h-screen bg-slate-950 px-6 py-10">
      <div className="mx-auto flex max-w-6xl flex-col gap-8">
        <header className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <p className="text-sm uppercase tracking-[0.25em] text-slate-400">Deployment Console</p>
            <h1 className="text-3xl font-semibold text-white">Agentless software delivery</h1>
          </div>
          <div className="flex items-center gap-3">
            <div className="rounded-full border border-slate-800 bg-slate-900/60 px-4 py-2 text-xs text-slate-300">
              Secure-first fallback enabled
            </div>
            <label className="flex items-center gap-2 rounded-full border border-slate-800 bg-slate-900/60 px-3 py-2 text-xs text-slate-300">
              <span>Demo mode</span>
              <button
                type="button"
                aria-pressed={isDemoMode}
                aria-label="Toggle demo mode"
                onClick={() => setIsDemoMode((value) => !value)}
                className={getToggleClass(isDemoMode)}
              >
                <span className={getKnobClass(isDemoMode)} />
              </button>
            </label>
          </div>
        </header>

        <nav className="flex flex-wrap items-center gap-3 text-sm text-slate-300">
          <button
            type="button"
            onClick={() => setPage('dashboard')}
            className={getNavClass(page === 'dashboard')}
          >
            Dashboard
          </button>
          <button
            type="button"
            onClick={() => setPage('deployments')}
            className={getNavClass(page === 'deployments')}
          >
            Deployments
          </button>
        </nav>

        {page === 'dashboard' ? (
          <div key={refreshKey}>
            <Dashboard isDemoMode={isDemoMode} onViewDeployments={handleViewDeployments} />
          </div>
        ) : page === 'deployments' ? (
          <Deployments
            isDemoMode={isDemoMode}
            statusFilter={deploymentFilter}
            onCreateDeployment={() => setPage('deploy')}
          />
        ) : (
          <Wizard onTaskCreated={handleTaskCreated} isDemoMode={isDemoMode} />
        )}
      </div>
    </div>
  )
}

function getNavClass(isActive: boolean) {
  if (isActive) return 'rounded-lg border border-indigo-400 bg-indigo-500/20 px-4 py-2 text-sm text-indigo-200'
  return 'rounded-lg border border-slate-700 bg-slate-900/60 px-4 py-2 text-sm text-slate-300'
}

function getToggleClass(isActive: boolean) {
  if (isActive) return 'relative h-5 w-10 rounded-full bg-emerald-500 transition'
  return 'relative h-5 w-10 rounded-full bg-slate-700 transition'
}

function getKnobClass(isActive: boolean) {
  if (isActive) return 'absolute left-0 top-0 h-5 w-5 translate-x-5 rounded-full bg-white shadow transition'
  return 'absolute left-0 top-0 h-5 w-5 translate-x-0 rounded-full bg-white shadow transition'
}
