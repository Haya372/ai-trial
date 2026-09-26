import { zodResolver } from '@hookform/resolvers/zod'
import { useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import type { CreateEventRequest, EventResponse } from '../../../api/generated'
import { createEvent } from '../../../api/generated'
import { runEventMutation } from '../runEventMutation'
import { type QuickRegistrationValues, quickRegistrationSchema } from '../types'
import { DEFAULT_EVENT_DURATION_MS } from '../utils'

export type QuickRegistrationPhase =
  | { status: 'input' }
  | { status: 'success'; event: EventResponse }

export function useQuickRegistrationForm(start: Date) {
  const queryClient = useQueryClient()
  const [phase, setPhase] = useState<QuickRegistrationPhase>({
    status: 'input',
  })
  const form = useForm<QuickRegistrationValues>({
    resolver: zodResolver(quickRegistrationSchema),
    mode: 'onTouched',
    defaultValues: { title: '' },
  })

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
      successMessage: '予定を登録しました',
      onSuccess: (createdEvent) => {
        setPhase({ status: 'success', event: createdEvent as EventResponse })
      },
    })
  }

  return { form, phase, onSubmit }
}
