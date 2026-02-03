import { useEffect, useState } from 'react'
import { getAssessments, getErrorCatalog, getMetrics, listTasks, type Assessment, type ErrorCatalogItem, type MetricsSummary, type Task } from '../api/client'

interface DashboardData {
  metrics: MetricsSummary | null
  errors: ErrorCatalogItem[]
  tasks: Task[]
  assessments: Assessment[]
  isLoading: boolean
  errorMessage: string | null
}

interface UseDashboardDataOptions {
  isDemoMode: boolean
}

export function useDashboardData({ isDemoMode }: UseDashboardDataOptions): DashboardData {
  const [metrics, setMetrics] = useState<MetricsSummary | null>(null)
  const [errors, setErrors] = useState<ErrorCatalogItem[]>([])
  const [tasks, setTasks] = useState<Task[]>([])
  const [assessments, setAssessments] = useState<Assessment[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)

  useEffect(() => {
    let isMounted = true

    if (isDemoMode) {
      setMetrics(demoMetrics)
      setErrors(demoErrors)
      setTasks(demoTasks)
      setAssessments(demoAssessments)
      setIsLoading(false)
      setErrorMessage(null)
      return () => {
        isMounted = false
      }
    }

    async function loadData() {
      try {
        const [metricsResponse, errorsResponse, tasksResponse, assessmentResponse] = await Promise.all([
          getMetrics(),
          getErrorCatalog(),
          listTasks(),
          getAssessments()
        ])

        if (!isMounted) return
        setMetrics(metricsResponse)
        setErrors(errorsResponse)
        setTasks(tasksResponse)
        setAssessments(assessmentResponse)
        setErrorMessage(null)
      } catch (error) {
        if (!isMounted) return
        setErrorMessage(error instanceof Error ? error.message : 'Failed to load data')
      } finally {
        if (!isMounted) return
        setIsLoading(false)
      }
    }

    loadData()
    const interval = window.setInterval(loadData, 5000)

    return () => {
      isMounted = false
      window.clearInterval(interval)
    }
  }, [isDemoMode])

  return {
    metrics,
    errors,
    tasks,
    assessments,
    isLoading,
    errorMessage
  }
}

const demoMetrics: MetricsSummary = {
  totalTasks: 18,
  runningTasks: 2,
  successTasks: 13,
  failedTasks: 3,
  pendingTasks: 3,
  targetsTotal: 220,
  targetsScanned: 198,
  failureReasons: {
    auth_denied: 2,
    port_closed: 1
  },
  successByOS: {
    linux: 8,
    windows: 4,
    macos: 1
  },
  failureByOS: {
    windows: 2,
    linux: 1
  },
  authMethods: {
    ssh_key: 9,
    winrm_https_cert: 3,
    winrm_https_userpass: 1
  }
}

const demoErrors: ErrorCatalogItem[] = [
  {
    code: 'auth_denied',
    message: 'Authentication denied',
    remediation: 'Ensure key access and verify target-side authorization policies.',
    steps: [
      'Verify the account or key has access on the target.',
      'Check the target auth policy and permissions.',
      'Retry with key-based auth first, then password if needed.'
    ]
  },
  {
    code: 'port_closed',
    message: 'Required port closed',
    remediation: 'Enable SSH or WinRM and restrict access to the controller subnet.',
    steps: [
      'Enable SSH (22) or WinRM HTTPS (5986) on the target.',
      'Restrict access to the controller subnet.',
      'Re-run the assessment scan.'
    ]
  }
]

const demoTasks: Task[] = [
  {
    id: 'demo-1',
    name: 'February rollout',
    status: 'running',
    targetCount: 120,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString()
  }
]

const demoAssessments: Assessment[] = [
  {
    targetId: 'demo-target-1',
    label: 'core-db-01',
    os: 'linux',
    reachable: true,
    openPorts: [22],
    predictedSuccess: 92,
    secureMethod: 'SSH key authentication',
    guidelines: ['Restrict SSH access to the controller subnet.', 'Use short-lived SSH keys.'],
    scannedAt: '2026-02-02 10:45:00 PST'
  },
  {
    targetId: 'demo-target-2',
    label: 'win-edge-04',
    os: 'windows',
    reachable: true,
    openPorts: [5986],
    predictedSuccess: 86,
    secureMethod: 'WinRM HTTPS (certificate)',
    guidelines: ['Ensure certificate chain is trusted.', 'Limit WinRM to HTTPS only.'],
    scannedAt: '2026-02-02 10:44:12 PST'
  }
]
