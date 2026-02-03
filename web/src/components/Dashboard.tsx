import { lazy, Suspense } from 'react'
import { useDashboardData } from '../state/use-dashboard-data'

const DonutChart = lazy(() =>
  import('./charts/DonutChart').then((module) => ({ default: module.DonutChart }))
)
const BarChart = lazy(() =>
  import('./charts/BarChart').then((module) => ({ default: module.BarChart }))
)

interface DashboardProps {
  isDemoMode: boolean
  onViewDeployments: (filter: 'success' | 'failed') => void
}

export function Dashboard({ isDemoMode, onViewDeployments }: DashboardProps) {
  const { metrics, errors, assessments, isLoading, errorMessage } = useDashboardData({ isDemoMode })

  if (isLoading) {
    return <div className="rounded-2xl border border-slate-800 bg-slate-900/60 p-6">Loading metrics...</div>
  }

  if (errorMessage || !metrics) {
    return (
      <div className="rounded-2xl border border-slate-800 bg-slate-900/60 p-6 text-rose-300">
        {errorMessage ?? 'Unable to load metrics'}
      </div>
    )
  }

  const failureReasons = Object.entries(metrics.failureReasons).map(([label, value]) => ({
    label,
    value
  }))
  const authMethods = Object.entries(metrics.authMethods).map(([label, value]) => ({
    label,
    value
  }))
  const successByOS = Object.entries(metrics.successByOS).map(([label, value]) => ({
    label,
    value
  }))

  return (
    <section className="rounded-2xl border border-slate-800 bg-slate-900/60 p-6 shadow-xl">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm uppercase tracking-[0.25em] text-slate-400">Live Metrics</p>
          <h2 className="text-2xl font-semibold text-white">Deployment performance</h2>
        </div>
        <div className="rounded-full bg-slate-800 px-4 py-2 text-xs text-slate-300">
          Active tasks: {metrics.runningTasks}
        </div>
      </div>

      <div className="mt-6 grid gap-4 lg:grid-cols-3">
        <button
          type="button"
          onClick={() => onViewDeployments('success')}
          className="text-left"
          aria-label="View successful task deployments"
        >
          <Suspense fallback={<div className="rounded-xl border border-slate-800 bg-slate-950/60 p-4">Loading chart...</div>}>
            <DonutChart
              value={metrics.successTasks}
              total={metrics.totalTasks}
              label="Successful tasks"
              color="rgba(34, 197, 94, 0.8)"
            />
          </Suspense>
        </button>
        <button
          type="button"
          onClick={() => onViewDeployments('failed')}
          className="text-left"
          aria-label="View failed task deployments"
        >
          <Suspense fallback={<div className="rounded-xl border border-slate-800 bg-slate-950/60 p-4">Loading chart...</div>}>
            <DonutChart
              value={metrics.failedTasks}
              total={metrics.totalTasks}
              label="Failed tasks"
              color="rgba(248, 113, 113, 0.8)"
            />
          </Suspense>
        </button>
        <Suspense fallback={<div className="rounded-xl border border-slate-800 bg-slate-950/60 p-4">Loading chart...</div>}>
          <DonutChart
            value={metrics.targetsScanned}
            total={metrics.targetsTotal}
            label="Coverage"
            color="rgba(96, 165, 250, 0.8)"
          />
        </Suspense>
      </div>

      <div className="mt-6 grid gap-4 lg:grid-cols-2">
        <div>
          <h3 className="text-sm font-semibold text-slate-200">Failure reasons</h3>
          <Suspense fallback={<div className="rounded-xl border border-slate-800 bg-slate-950/60 p-4">Loading chart...</div>}>
            <BarChart
              data={failureReasons.length ? failureReasons : [{ label: 'No failures', value: 0 }]}
              color="rgba(248, 113, 113, 0.7)"
            />
          </Suspense>
        </div>
        <div className="rounded-xl border border-slate-800 bg-slate-950/60 p-4">
          <h3 className="text-sm font-semibold text-slate-200">Recommended actions</h3>
          <ul className="mt-3 space-y-3 text-sm text-slate-300">
            {errors.map((error) => (
              <li key={error.code} className="rounded-lg border border-slate-800 bg-slate-900/80 p-3">
                <p className="font-semibold text-slate-200">{error.message}</p>
                <p className="text-xs text-slate-400">{error.remediation}</p>
                <ol className="mt-2 list-decimal space-y-1 pl-5 text-xs text-slate-300">
                  {error.steps?.map((step) => (
                    <li key={step}>{step}</li>
                  ))}
                </ol>
              </li>
            ))}
          </ul>
        </div>
      </div>

      <div className="mt-6 grid gap-4 lg:grid-cols-2">
        <div>
          <h3 className="text-sm font-semibold text-slate-200">Success by OS</h3>
          <Suspense fallback={<div className="rounded-xl border border-slate-800 bg-slate-950/60 p-4">Loading chart...</div>}>
            <BarChart
              data={successByOS.length ? successByOS : [{ label: 'No data', value: 0 }]}
              color="rgba(34, 197, 94, 0.7)"
            />
          </Suspense>
        </div>
        <div>
          <h3 className="text-sm font-semibold text-slate-200">Auth method usage</h3>
          <Suspense fallback={<div className="rounded-xl border border-slate-800 bg-slate-950/60 p-4">Loading chart...</div>}>
            <BarChart
              data={authMethods.length ? authMethods : [{ label: 'No data', value: 0 }]}
              color="rgba(96, 165, 250, 0.7)"
            />
          </Suspense>
        </div>
      </div>

      <div className="mt-6 rounded-xl border border-slate-800 bg-slate-950/60 p-4">
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-semibold text-slate-200">Pre-deployment assessment</h3>
          <span className="text-xs text-slate-400">{isDemoMode ? 'Demo dataset' : 'Secure-first guidance'}</span>
        </div>
        <div className="mt-4 grid gap-3">
          {assessments.length === 0 ? (
            <div className="text-sm text-slate-400">No target assessments yet.</div>
          ) : (
            assessments.map((assessment) => (
              <div key={assessment.targetId} className="rounded-lg border border-slate-800 bg-slate-900/80 p-3">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <div>
                    <p className="text-sm font-semibold text-slate-200">{assessment.label}</p>
                    <p className="text-xs text-slate-400">
                      OS: {assessment.os} · Last scan: {assessment.scannedAt || 'Not scanned'}
                    </p>
                  </div>
                  <div className="rounded-full bg-slate-800 px-3 py-1 text-xs text-slate-300">
                    Predicted success: {assessment.predictedSuccess}%
                  </div>
                </div>
                <div className="mt-3 grid gap-2 md:grid-cols-2">
                  <div>
                    <p className="text-xs text-slate-400">Secure method</p>
                    <p className="text-sm text-slate-200">{assessment.secureMethod}</p>
                  </div>
                  <div>
                    <p className="text-xs text-slate-400">Guidelines</p>
                    <ul className="text-xs text-slate-300">
                      {assessment.guidelines.map((guideline) => (
                        <li key={guideline}>• {guideline}</li>
                      ))}
                    </ul>
                  </div>
                </div>
              </div>
            ))
          )}
        </div>
      </div>

    </section>
  )
}

