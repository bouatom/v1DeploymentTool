import { z } from 'zod'

export const wizardSchema = z
  .object({
    taskName: z.string().min(3, 'Task name is required'),
    targets: z.string().min(3, 'Targets are required'),
    aggressiveness: z.number().min(1).max(5),
    sshUsername: z.string().optional(),
    sshPassword: z.string().optional(),
    sshPrivateKey: z.string().optional(),
    winrmUsername: z.string().optional(),
    winrmPassword: z.string().optional(),
    binaryUrl: z.string().url('Installer URL is required'),
    checksum: z.string().optional(),
    installerId: z.string().optional(),
    scheduleType: z.enum(['now', 'later']),
    startAt: z.string().optional()
  })
  .refine(
    (values) => (values.scheduleType === 'later' ? Boolean(values.startAt) : true),
    { path: ['startAt'], message: 'Start time is required when scheduling later' }
  )
  .refine(
    (values) => (values.scheduleType === 'later' ? Boolean(values.startAt) : true),
    { path: ['startAt'], message: 'Start time is required when scheduling later' }
  )

export type WizardFormValues = z.infer<typeof wizardSchema>
