const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'
const API_KEY = import.meta.env.VITE_API_KEY ?? ''

export interface Task {
  id: string
  name: string
  status: string
  targetCount: number
  createdAt: string
  updatedAt: string
}

export interface MetricsSummary {
  totalTasks: number
  runningTasks: number
  successTasks: number
  failedTasks: number
  pendingTasks: number
  targetsTotal: number
  targetsScanned: number
  failureReasons: Record<string, number>
  successByOS: Record<string, number>
  failureByOS: Record<string, number>
  authMethods: Record<string, number>
}

export interface ErrorCatalogItem {
  code: string
  message: string
  remediation: string
  steps: string[]
}

export interface Assessment {
  targetId: string
  label: string
  os: string
  reachable: boolean | null
  openPorts: number[]
  predictedSuccess: number
  secureMethod: string
  guidelines: string[]
  scannedAt: string
}

export interface TaskDeploymentDetail {
  id: string
  taskRunId: string
  targetId: string
  targetLabel: string
  targetOS: string
  status: string
  authMethod: string
  errorCode: string
  errorMessage: string
  remediation: string
  finishedAt: string
}

export interface CreateTaskInput {
  name: string
  targetCount: number
}

export interface ScanInput {
  targetCount: number
  targetsScanned: number
}

export interface ExecuteScanInput {
  targets: string[]
  aggressiveness: number
}

export interface ExecuteScanResponse {
  targetCount: number
  targetsScanned: number
  errors: string[]
}

export interface UpdateRunInput {
  runId: string
  status: string
}

export async function createTask(input: CreateTaskInput) {
  return requestJson<Task>({
    path: '/api/tasks',
    method: 'POST',
    body: input
  })
}

export async function listTasks() {
  return requestJson<Task[]>({
    path: '/api/tasks',
    method: 'GET'
  })
}

export async function createRun(taskId: string) {
  return requestJson<{ id: string }>({
    path: `/api/tasks/${taskId}/runs`,
    method: 'POST'
  })
}

export async function updateRun(input: UpdateRunInput) {
  return requestJson<{ id: string }>({
    path: `/api/runs/${input.runId}`,
    method: 'PATCH',
    body: { status: input.status }
  })
}

export async function recordScan(input: ScanInput) {
  return requestJson({
    path: '/api/scans',
    method: 'POST',
    body: input
  })
}

export async function getMetrics() {
  return requestJson<MetricsSummary>({
    path: '/api/metrics',
    method: 'GET'
  })
}

export async function getErrorCatalog() {
  return requestJson<ErrorCatalogItem[]>({
    path: '/api/errors',
    method: 'GET'
  })
}

export async function getAssessments() {
  return requestJson<Assessment[]>({
    path: '/api/assessments',
    method: 'GET'
  })
}

export async function listTaskDeployments(taskId: string) {
  return requestJson<TaskDeploymentDetail[]>({
    path: `/api/tasks/${taskId}/deployments`,
    method: 'GET'
  })
}

export async function downloadTaskReport(taskId: string, format: 'csv' | 'pdf') {
  const response = await fetch(`${API_BASE_URL}/api/tasks/${taskId}/exports/${format}`, {
    method: 'GET',
    headers: {
      ...(API_KEY ? { 'X-API-Key': API_KEY } : {})
    }
  })

  if (!response.ok) {
    const message = await response.text()
    throw new Error(message || 'Export failed')
  }

  const blob = await response.blob()
  return blob
}

export async function executeScan(input: ExecuteScanInput) {
  return requestJson<ExecuteScanResponse>({
    path: '/api/scans/execute',
    method: 'POST',
    body: input
  })
}

export async function uploadInstaller(file: File) {
  const formData = new FormData()
  formData.append('file', file)

  const response = await fetch(`${API_BASE_URL}/api/uploads/installer`, {
    method: 'POST',
    headers: {
      ...(API_KEY ? { 'X-API-Key': API_KEY } : {})
    },
    body: formData
  })

  if (!response.ok) {
    const message = await response.text()
    throw new Error(message || 'Upload failed')
  }

  return response.json() as Promise<{
    url: string
    filename: string
    checksum: string
    installerId: string
    packageType: string
    osFamily: string
  }>
}

export function getApiBaseUrl() {
  return API_BASE_URL
}

interface RequestParams {
  path: string
  method: 'GET' | 'POST' | 'PATCH'
  body?: unknown
}

async function requestJson<T>({ path, method, body }: RequestParams): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    method,
    headers: {
      'Content-Type': 'application/json',
      ...(API_KEY ? { 'X-API-Key': API_KEY } : {})
    },
    body: body ? JSON.stringify(body) : undefined
  })

  if (!response.ok) {
    const message = await response.text()
    throw new Error(message || 'Request failed')
  }

  return response.json() as Promise<T>
}
