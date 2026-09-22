import { zodResolver } from '@hookform/resolvers/zod'
import { useQueryClient } from '@tanstack/react-query'
import { toast } from '@repo/ui'
import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import type { CreateEventRequest, EventResponse } from '../../../api/generated'
import { createEvent, updateEvent } from '../../../api/generated'
import { eventsKeys } from '../../../lib/queryKeys'
import {
  type EventFormMode,
  type EventFormValues,
  eventFormSchema,
} from '../types'
import { getEventErrorMessage, toFormValues, toIsoString } from '../utils'

function toRequestPayload(data: EventFormValues): CreateEventRequest {
  return {
    title: data.title,
    startAt: toIsoString(data.startAt),
    endAt: toIsoString(data.endAt),
    description: data.description || null,
    location: data.location || null,
    url: data.url || null,
  }
}

export function useEventForm(
  mode: EventFormMode,
  event: EventResponse | null,
  onSaved: () => void,
) {
  const queryClient = useQueryClient()
  const form = useForm<EventFormValues>({
    resolver: zodResolver(eventFormSchema),
    mode: 'onTouched',
    defaultValues: toFormValues(event),
  })

  useEffect(() => {
    form.reset(toFormValues(event))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [event, mode])

  const onSubmit = async (data: EventFormValues) => {
    try {
      const payload = toRequestPayload(data)
      let res: Awaited<ReturnType<typeof createEvent | typeof updateEvent>>
      if (mode === 'create') {
        res = await createEvent(payload)
      } else {
        if (!event) return
        res = await updateEvent(event.id, payload)
      }
      const expectedStatus = mode === 'create' ? 201 : 200
      if (res.status !== expectedStatus) {
        toast.error(getEventErrorMessage(res.data, mode))
        return
      }
      await queryClient.invalidateQueries({ queryKey: eventsKeys.all })
      toast.success(
        mode === 'create' ? '予定を登録しました' : '予定を更新しました',
      )
      onSaved()
    } catch (error) {
      toast.error(getEventErrorMessage(error, mode))
    }
  }

  return { form, onSubmit }
}
