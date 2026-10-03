import type { TFunction } from 'i18next'
import { z } from 'zod'
import type { EventResponse } from '../../api/generated'

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

export function createShareLinkFormSchema(
  t: TFunction<'eventshare'>,
  event: EventResponse | null,
) {
  return z
    .object({
      expiresAt: z.string().min(1, t('validation.expiresAtRequired')),
    })
    .refine((data) => Number.isFinite(new Date(data.expiresAt).getTime()), {
      message: t('validation.expiresAtRequired'),
      path: ['expiresAt'],
    })
    .refine((data) => new Date(data.expiresAt).getTime() > Date.now(), {
      message: t('validation.expiresAtInFuture'),
      path: ['expiresAt'],
    })
    .refine(
      (data) => !event || new Date(data.expiresAt) >= new Date(event.startAt),
      { message: t('validation.expiresAtAfterStart'), path: ['expiresAt'] },
    )
}

export type ShareLinkFormValues = z.infer<
  ReturnType<typeof createShareLinkFormSchema>
>
