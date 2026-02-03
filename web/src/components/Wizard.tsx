import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { createRun, createTask, executeScan, getApiBaseUrl, uploadInstaller } from '../api/client'
import { wizardSchema, type WizardFormValues } from './wizard/schema'

interface WizardProps {
  onTaskCreated: () => void
  isDemoMode: boolean
}

const steps = ['Targets', 'Credentials', 'Installer', 'Schedule', 'Review']

export function Wizard({ onTaskCreated, isDemoMode }: WizardProps) {
  const [stepIndex, setStepIndex] = useState(0)
  const [statusMessage, setStatusMessage] = useState<string | null>(null)
  const [scanStatus, setScanStatus] = useState<string | null>(null)
  const [scanResult, setScanResult] = useState<string | null>(null)
  const [installerStatus, setInstallerStatus] = useState<string | null>(null)
  const [installerFileName, setInstallerFileName] = useState<string | null>(null)
  const [installerMeta, setInstallerMeta] = useState<{ packageType: string; osFamily: string } | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [isScanning, setIsScanning] = useState(false)
  const [isUploadingInstaller, setIsUploadingInstaller] = useState(false)

  const form = useForm<WizardFormValues>({
    resolver: zodResolver(wizardSchema),
    defaultValues: {
      taskName: '',
      targets: '',
      aggressiveness: 3,
      scheduleType: 'now'
    },
    mode: 'onBlur'
  })

  const handleNext = async () => {
    const isStepValid = await form.trigger(getStepFields(stepIndex))
    if (!isStepValid) return
    setStepIndex((current) => Math.min(current + 1, steps.length - 1))
  }

  const handleBack = () => {
    setStepIndex((current) => Math.max(current - 1, 0))
  }

  const handleSubmit = form.handleSubmit(async (values) => {
    if (isDemoMode) {
      setStatusMessage('Demo mode: task created (simulation)')
      onTaskCreated()
      return
    }

    setIsSubmitting(true)
    setStatusMessage(null)

    try {
      const targetCount = parseTargets(values.targets).length
      const task = await createTask({ name: values.taskName, targetCount })
      await createRun(task.id)
      setStatusMessage('Task created. Installer settings saved for deployment execution.')
      onTaskCreated()
      setStepIndex(0)
      form.reset()
    } catch (error) {
      setStatusMessage(error instanceof Error ? error.message : 'Failed to create task')
    } finally {
      setIsSubmitting(false)
    }
  })

  const handleScan = async () => {
    if (isDemoMode) {
      setScanStatus('Demo mode: assessment scan simulated')
      setScanResult('198 scanned, 2 issues')
      return
    }

    setIsScanning(true)
    setScanStatus('Running assessment scan...')
    setScanResult(null)

    try {
      const targets = parseTargets(form.getValues('targets'))
      const response = await executeScan({
        targets,
        aggressiveness: form.getValues('aggressiveness')
      })
      setScanStatus('Assessment scan complete')
      setScanResult(`${response.targetsScanned} scanned, ${response.errors.length} issues`)
    } catch (error) {
      setScanStatus('Assessment scan failed')
      setScanResult(error instanceof Error ? error.message : 'Scan failed')
    } finally {
      setIsScanning(false)
    }
  }

  const handleInstallerFileChange = async (file: File | null) => {
    if (!file) return
    if (isDemoMode) {
      const demoUrl = `${getApiBaseUrl()}/uploads/demo-installer.bin`
      form.setValue('binaryUrl', demoUrl, { shouldValidate: true })
      form.setValue('checksum', 'demo-checksum', { shouldValidate: false })
      form.setValue('installerId', 'demo-installer', { shouldValidate: false })
      setInstallerFileName(file.name)
      setInstallerMeta({ packageType: 'binary', osFamily: 'any' })
      setInstallerStatus('Demo mode: installer staged')
      return
    }

    setIsUploadingInstaller(true)
    setInstallerStatus('Uploading installer...')
    setInstallerFileName(file.name)

    try {
      const response = await uploadInstaller(file)
      form.setValue('binaryUrl', response.url, { shouldValidate: true })
      form.setValue('checksum', response.checksum, { shouldValidate: false })
      form.setValue('installerId', response.installerId, { shouldValidate: false })
      setInstallerMeta({ packageType: response.packageType, osFamily: response.osFamily })
      setInstallerStatus('Installer uploaded and ready')
    } catch (error) {
      setInstallerStatus(error instanceof Error ? error.message : 'Installer upload failed')
    } finally {
      setIsUploadingInstaller(false)
    }
  }

  return (
    <section className="rounded-2xl border border-slate-800 bg-slate-900/60 p-6 shadow-xl">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <p className="text-sm uppercase tracking-[0.25em] text-slate-400">Deployment Wizard</p>
          <h2 className="text-2xl font-semibold text-white">Create a deployment task</h2>
        </div>
        <div className="flex items-center gap-2 text-xs text-slate-400">
          {steps.map((label, index) => (
            <span
              key={label}
              className={getStepClass(index === stepIndex)}
            >
              {index + 1}. {label}
            </span>
          ))}
        </div>
      </div>

      <form onSubmit={handleSubmit} className="mt-6 space-y-6">
        {stepIndex === 0 && (
          <div className="space-y-4">
            <div>
              <label className="text-sm text-slate-300" htmlFor="taskName">
                Task name
              </label>
              <input
                id="taskName"
                {...form.register('taskName')}
                className="mt-2 w-full rounded-lg border border-slate-700 bg-slate-950 px-4 py-2 text-sm text-white"
                placeholder="Endpoint rollout - February"
              />
              <p className="mt-1 text-xs text-rose-300">{form.formState.errors.taskName?.message}</p>
            </div>
            <div>
              <label className="text-sm text-slate-300" htmlFor="targets">
                Targets (hostnames, IPs, CIDR)
              </label>
              <textarea
                id="targets"
                {...form.register('targets')}
                className="mt-2 h-28 w-full rounded-lg border border-slate-700 bg-slate-950 px-4 py-2 text-sm text-white"
                placeholder="host-1.local, 192.168.0.10, 10.0.0.0/24"
              />
              <p className="mt-1 text-xs text-rose-300">{form.formState.errors.targets?.message}</p>
            </div>
            <div>
              <label className="text-sm text-slate-300" htmlFor="aggressiveness">
                Scan aggressiveness (1-5)
              </label>
              <input
                id="aggressiveness"
                type="range"
                min={1}
                max={5}
                step={1}
                {...form.register('aggressiveness', { valueAsNumber: true })}
                className="mt-3 w-full"
              />
              <p className="text-xs text-slate-500">
                Lower values reduce network impact. Higher values speed discovery.
              </p>
            </div>
          </div>
        )}

        {stepIndex === 1 && (
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-4 rounded-xl border border-slate-800 bg-slate-950/60 p-4">
              <h3 className="text-sm font-semibold text-slate-200">Linux / macOS (SSH)</h3>
              <input
                {...form.register('sshUsername')}
                className="w-full rounded-lg border border-slate-700 bg-slate-950 px-4 py-2 text-sm text-white"
                placeholder="SSH username"
              />
              <textarea
                {...form.register('sshPrivateKey')}
                className="h-24 w-full rounded-lg border border-slate-700 bg-slate-950 px-4 py-2 text-xs text-white"
                placeholder="Paste OpenSSH private key (-----BEGIN OPENSSH PRIVATE KEY-----)"
              />
              <input
                {...form.register('sshPassword')}
                type="password"
                className="w-full rounded-lg border border-slate-700 bg-slate-950 px-4 py-2 text-sm text-white"
                placeholder="SSH password (fallback)"
              />
              <div className="text-xs text-slate-500">
                <p>Default order: SSH key, then password.</p>
                <p>Example key format: OpenSSH PEM (starts with BEGIN OPENSSH PRIVATE KEY).</p>
              </div>
            </div>
            <div className="space-y-4 rounded-xl border border-slate-800 bg-slate-950/60 p-4">
              <h3 className="text-sm font-semibold text-slate-200">Windows (WinRM HTTPS)</h3>
              <input
                {...form.register('winrmUsername')}
                className="w-full rounded-lg border border-slate-700 bg-slate-950 px-4 py-2 text-sm text-white"
                placeholder="WinRM username"
              />
              <input
                {...form.register('winrmPassword')}
                type="password"
                className="w-full rounded-lg border border-slate-700 bg-slate-950 px-4 py-2 text-sm text-white"
                placeholder="WinRM password (fallback)"
              />
              <div className="text-xs text-slate-500">
                <p>Default order: certificate, then user/password.</p>
                <p>Example usernames: DOMAIN\\admin-user or admin-user@host.</p>
              </div>
            </div>
          </div>
        )}

        {stepIndex === 2 && (
          <div className="space-y-4">
            <div>
              <label className="text-sm text-slate-300" htmlFor="binaryUrl">
                Installer URL
              </label>
              <input
                id="binaryUrl"
                {...form.register('binaryUrl')}
                className="mt-2 w-full rounded-lg border border-slate-700 bg-slate-950 px-4 py-2 text-sm text-white"
                placeholder="https://example.com/installer.bin"
              />
              <p className="mt-1 text-xs text-rose-300">{form.formState.errors.binaryUrl?.message}</p>
              <p className="text-xs text-slate-500">
                Provide a direct HTTPS URL reachable from the target machines.
              </p>
              <p className="text-xs text-slate-500">
                Avoid localhost URLs unless the targets run on the same host as the controller.
              </p>
            </div>
            <div>
              <label className="text-sm text-slate-300" htmlFor="checksum">
                SHA256 checksum (optional)
              </label>
              <input
                id="checksum"
                {...form.register('checksum')}
                className="mt-2 w-full rounded-lg border border-slate-700 bg-slate-950 px-4 py-2 text-sm text-white"
                placeholder="auto-filled after upload or paste hash"
              />
            </div>
            <div className="rounded-lg border border-slate-800 bg-slate-950/60 p-3">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <p className="text-sm text-slate-200">Installer file</p>
                  <p className="text-xs text-slate-500">
                    Select an installer binary from your machine. Create separate tasks for macOS, Linux, and Windows installers.
                  </p>
                  {installerFileName ? (
                    <p className="text-xs text-slate-400">Selected: {installerFileName}</p>
                  ) : null}
                  {installerStatus ? <p className="text-xs text-slate-400">{installerStatus}</p> : null}
                </div>
                <label className="cursor-pointer rounded-lg border border-slate-700 px-3 py-2 text-xs text-slate-200">
                  {isUploadingInstaller ? 'Uploading...' : 'Select file'}
                  <input
                    type="file"
                    className="hidden"
                    onChange={(event) => handleInstallerFileChange(event.target.files?.[0] ?? null)}
                  />
                </label>
              </div>
              <p className="mt-2 text-xs text-slate-500">
                The uploaded file will be served from the controller for target downloads.
              </p>
            </div>
            <div className="rounded-lg border border-slate-800 bg-slate-950/60 p-3 text-xs text-slate-400">
              <p className="font-semibold text-slate-200">Detected package</p>
              <p className="mt-1">
                {installerMeta ? `${installerMeta.osFamily} / ${installerMeta.packageType}` : 'No package metadata yet.'}
              </p>
            </div>
            <div className="rounded-lg border border-slate-800 bg-slate-950/60 p-3 text-xs text-slate-400">
              Destination path is selected automatically based on the target OS. Windows uses
              <span className="ml-1 text-slate-200">C:\\V1SGDeploymentTool</span> and Linux/macOS use
              <span className="ml-1 text-slate-200">/tmp/V1SGDeploymentTool</span>.
              The folder is removed after install completes.
            </div>
          </div>
        )}

        {stepIndex === 3 && (
          <div className="space-y-4">
            <div className="flex gap-3">
              <button
                type="button"
                onClick={() => form.setValue('scheduleType', 'now')}
                className={getScheduleClass(form.watch('scheduleType') === 'now')}
              >
                Run now
              </button>
              <button
                type="button"
                onClick={() => form.setValue('scheduleType', 'later')}
                className={getScheduleClass(form.watch('scheduleType') === 'later')}
              >
                Schedule
              </button>
            </div>
            {form.watch('scheduleType') === 'later' && (
              <div>
                <label className="text-sm text-slate-300" htmlFor="startAt">
                  Start time
                </label>
                <input
                  id="startAt"
                  type="datetime-local"
                  {...form.register('startAt')}
                  className="mt-2 w-full rounded-lg border border-slate-700 bg-slate-950 px-4 py-2 text-sm text-white"
                />
                <p className="mt-1 text-xs text-rose-300">{form.formState.errors.startAt?.message}</p>
              </div>
            )}
            <p className="text-xs text-slate-500">
              Concurrency and rate limits are applied automatically based on aggressiveness.
            </p>
          </div>
        )}

        {stepIndex === 4 && (
          <div className="space-y-4 rounded-xl border border-slate-800 bg-slate-950/60 p-4">
            <h3 className="text-sm font-semibold text-slate-200">Review</h3>
            <div className="text-sm text-slate-400">
              <p>
                <span className="text-slate-200">Task name:</span> {form.watch('taskName') || 'Untitled'}
              </p>
              <p>
                <span className="text-slate-200">Targets:</span> {parseTargets(form.watch('targets')).length}
              </p>
              <p>
                <span className="text-slate-200">Installer URL:</span> {form.watch('binaryUrl')}
              </p>
              <p>
                <span className="text-slate-200">Checksum:</span> {form.watch('checksum') || 'Not provided'}
              </p>
              <p>
                <span className="text-slate-200">Aggressiveness:</span> {form.watch('aggressiveness')}
              </p>
              <p>
                <span className="text-slate-200">Schedule:</span> {form.watch('scheduleType')}
              </p>
            </div>
            <p className="text-xs text-slate-500">
              The controller will detect OS and pick secure-first auth methods automatically.
            </p>
            <div className="rounded-lg border border-slate-800 bg-slate-900/60 p-3">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <div>
                  <p className="text-xs text-slate-400">Pre-deployment assessment</p>
                  <p className="text-sm text-slate-200">{scanStatus ?? 'Run a scan to assess readiness.'}</p>
                  {scanResult ? <p className="text-xs text-slate-400">{scanResult}</p> : null}
                </div>
                <button
                  type="button"
                  onClick={handleScan}
                  disabled={isScanning}
                  className="rounded-lg border border-slate-700 px-3 py-2 text-xs text-slate-200 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  {isScanning ? 'Scanning...' : 'Run assessment'}
                </button>
              </div>
            </div>
          </div>
        )}

        <div className="flex flex-wrap items-center justify-between gap-3 border-t border-slate-800 pt-4">
          <div className="text-sm text-slate-400">{statusMessage}</div>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={handleBack}
              disabled={stepIndex === 0}
              className="rounded-lg border border-slate-700 px-4 py-2 text-sm text-slate-200 disabled:cursor-not-allowed disabled:opacity-50"
            >
              Back
            </button>
            {stepIndex < steps.length - 1 ? (
              <button
                type="button"
                onClick={handleNext}
                className="rounded-lg bg-indigo-500 px-4 py-2 text-sm font-semibold text-white"
              >
                Continue
              </button>
            ) : (
              <button
                type="submit"
                disabled={isSubmitting}
                className="rounded-lg bg-emerald-500 px-4 py-2 text-sm font-semibold text-white disabled:cursor-not-allowed disabled:opacity-50"
              >
                Create task
              </button>
            )}
          </div>
        </div>
      </form>
    </section>
  )
}

function parseTargets(targets: string) {
  return targets
    .split(/[\n,]+/)
    .map((value) => value.trim())
    .filter((value) => value.length > 0)
}

function getStepClass(isActive: boolean) {
  if (isActive) return 'rounded-full bg-indigo-500/30 px-3 py-1 text-indigo-200'
  return 'rounded-full bg-slate-800 px-3 py-1 text-slate-400'
}

function getScheduleClass(isActive: boolean) {
  if (isActive) return 'rounded-lg border border-indigo-400 bg-indigo-500/20 px-4 py-2 text-sm text-indigo-200'
  return 'rounded-lg border border-slate-700 px-4 py-2 text-sm text-slate-300'
}

function getStepFields(stepIndex: number): Array<keyof WizardFormValues> {
  if (stepIndex === 0) return ['taskName', 'targets', 'aggressiveness']
  if (stepIndex === 1) return ['sshUsername', 'sshPassword', 'sshPrivateKey', 'winrmUsername', 'winrmPassword']
  if (stepIndex === 2) return ['binaryUrl']
  if (stepIndex === 3) return ['scheduleType', 'startAt']
  return ['taskName', 'targets', 'binaryUrl']
}
