import { zodResolver } from '@hookform/resolvers/zod'
import { useQueryClient } from '@tanstack/react-query'
import { useEffect, useMemo, useState } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import type { CreateEventRequest, EventResponse } from '../../../api/generated'
import { createEvent } from '../../../api/generated'
import { runEventMutation } from '../runEventMutation'
import {
  createQuickRegistrationSchema,
  type QuickRegistrationValues,
} from '../types'
import { DEFAULT_EVENT_DURATION_MS } from '../utils'

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

  useEffect(() => {
    const erroredFields = Object.keys(form.formState.errors) as Array<
      keyof QuickRegistrationValues
    >
    if (erroredFields.length > 0) {
      form.trigger(erroredFields)
    }
    // Deliberately keyed on `t` alone: this re-validates currently-errored
    // fields only when the language changes, so a displayed error message
    // is re-translated instead of staying stuck until the field is next
    // touched. Adding form.formState.errors here would make trigger()
    // re-fire this same effect in a loop.
    // oxlint-disable-next-line react-hooks/exhaustive-deps
  }, [t])

  const onSubmit = async (data: QuickRegistrationValues) => {
    const endAt = new Date(start.getTime() + DEFAULT_EVENT_DURATION_MS)
    const payload: CreateEventRequest = {
      title: data.title,
      startAt: start.toISOString(),
      endAt: endAt.toISOString(),
      description: null,
      location: null,
      url: null,
    }
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
