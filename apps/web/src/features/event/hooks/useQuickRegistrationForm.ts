import { zodResolver } from '@hookform/resolvers/zod'
import { useQueryClient } from '@tanstack/react-query'
import { useMemo, useState } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import type { EventResponse } from '../../../api/generated'
import { createEvent } from '../../../api/generated'
import { useRevalidateOnLanguageChange } from '../../../hooks/useRevalidateOnLanguageChange'
import { runEventMutation } from '../runEventMutation'
import {
  createQuickRegistrationSchema,
  type QuickRegistrationValues,
} from '../types'
import { DEFAULT_EVENT_DURATION_MS, toCreateEventRequest } from '../utils'

export type QuickRegistrationPhase =
  | { status: 'input' }
  | { status: 'success'; event: EventResponse }

export function useQuickRegistrationForm(start: Date) {
  const { t } = useTranslation(['event', 'common'])
  const queryClient = useQueryClient()
  const [phase, setPhase] = useState<QuickRegistrationPhase>({
    status: 'input',
  })
  const schema = useMemo(() => createQuickRegistrationSchema(t), [t])
  const form = useForm<QuickRegistrationValues>({
    resolver: zodResolver(schema),
    mode: 'onTouched',
    defaultValues: { title: '' },
  })

  useRevalidateOnLanguageChange(form)

  const onSubmit = async (data: QuickRegistrationValues) => {
    const endAt = new Date(start.getTime() + DEFAULT_EVENT_DURATION_MS)
    const payload = toCreateEventRequest({
      title: data.title,
      startAt: start.toISOString(),
      endAt: endAt.toISOString(),
    })
    await runEventMutation({
      queryClient,
      request: createEvent(payload),
      expectedStatus: 201,
      mode: 'create',
      successMessage: t('toast.createSuccess'),
      t,
      onSuccess: (createdEvent) => {
        setPhase({ status: 'success', event: createdEvent as EventResponse })
      },
    })
  }

  return { form, phase, onSubmit }
}
