import { zodResolver } from '@hookform/resolvers/zod'
import { useQueryClient } from '@tanstack/react-query'
import { useEffect, useMemo } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import type { CreateEventRequest, EventResponse } from '../../../api/generated'
import { createEvent, updateEvent } from '../../../api/generated'
import { useRevalidateOnLanguageChange } from '../../../hooks/useRevalidateOnLanguageChange'
import { runEventMutation } from '../runEventMutation'
import {
  createEventFormSchema,
  type EventFormMode,
  type EventFormValues,
} from '../types'
import { toFormValues, toIsoString } from '../utils'

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
  open: boolean,
  mode: EventFormMode,
  event: EventResponse | null,
  initialStart: Date | null | undefined,
  onSaved: () => void,
) {
  const { t } = useTranslation(['event', 'common'])
  const queryClient = useQueryClient()
  const schema = useMemo(() => createEventFormSchema(t), [t])
  const form = useForm<EventFormValues>({
    resolver: zodResolver(schema),
    mode: 'onTouched',
    defaultValues: toFormValues(event, initialStart),
  })

  useEffect(() => {
    // open のたびにリセットする。event/mode が前回と同じ組み合わせで
    // 再オープンされた場合でも入力途中の値を破棄するため、
    // event/mode ではなく open のみを依存にする。
    if (open) form.reset(toFormValues(event, initialStart))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  useRevalidateOnLanguageChange(form)

  const onSubmit = async (data: EventFormValues) => {
    const payload = toRequestPayload(data)
    let request: ReturnType<typeof createEvent> | ReturnType<typeof updateEvent>
    if (mode === 'create') {
      request = createEvent(payload)
    } else {
      if (!event) return
      request = updateEvent(event.id, payload)
    }
    await runEventMutation({
      queryClient,
      request,
      expectedStatus: mode === 'create' ? 201 : 200,
      mode,
      successMessage:
        mode === 'create' ? t('toast.createSuccess') : t('toast.updateSuccess'),
      t,
      onSuccess: onSaved,
    })
  }

  return { form, onSubmit }
}
