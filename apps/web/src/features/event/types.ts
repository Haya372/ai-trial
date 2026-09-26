import type { TFunction } from 'i18next'
import { z } from 'zod'

export function createEventFormSchema(t: TFunction<'event'>) {
  return z
    .object({
      title: z.string().min(1, t('validation.titleRequired')),
      startAt: z.string().min(1, t('validation.startAtRequired')),
      endAt: z.string().min(1, t('validation.endAtRequired')),
      description: z.string().optional(),
      location: z.string().optional(),
      url: z.string().optional(),
    })
    .refine((data) => new Date(data.endAt) > new Date(data.startAt), {
      message: t('validation.endAtAfterStart'),
      path: ['endAt'],
    })
}

export type EventFormValues = z.infer<ReturnType<typeof createEventFormSchema>>
export type EventFormMode = 'create' | 'edit'

export function createQuickRegistrationSchema(t: TFunction<'event'>) {
  return z.object({
    title: z.string().min(1, t('validation.titleRequired')),
  })
}

export type QuickRegistrationValues = z.infer<
  ReturnType<typeof createQuickRegistrationSchema>
>
