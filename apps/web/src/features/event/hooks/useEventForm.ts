import { zodResolver } from '@hookform/resolvers/zod'
import { useQueryClient } from '@tanstack/react-query'
import { useEffect, useMemo } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import type { CreateEventRequest, EventResponse } from '../../../api/generated'
import { createEvent, updateEvent } from '../../../api/generated'
import { useRevalidateOnLanguageChange } from '../../../hooks/useRevalidateOnLanguageChange'
import { hasErrorCode } from '../../../lib/errorMessage'
import { tagsKeys } from '../../../lib/queryKeys'
import { runEventMutation } from '../runEventMutation'
import {
  createEventFormSchema,
  type EventFormMode,
  type EventFormValues,
} from '../types'
import { toCreateEventRequest, toFormValues, toIsoString } from '../utils'

function toRequestPayload(
  data: EventFormValues,
  includeTags: boolean,
): CreateEventRequest {
  const payload = toCreateEventRequest({
    title: data.title,
    startAt: toIsoString(data.startAt),
    endAt: toIsoString(data.endAt),
    description: data.description,
    location: data.location,
    url: data.url,
  })
  return includeTags ? { ...payload, tagIds: data.tagIds } : payload
}

function hasDeletedTagError(data: unknown): boolean {
  if (!hasErrorCode(data) || data.code !== 'VALIDATION_ERROR') return false
  const { details } = data as { details?: { field: string; code: string }[] }
  return (
    details?.some(
      (detail) => detail.field === 'tagIds' && detail.code === 'NOT_FOUND',
    ) ?? false
  )
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

  // Events added via a share link are read-only, so their tags are never sent.
  const tagsEditable = !(mode === 'edit' && event?.isSubscribed)

  // A selected tag may have been deleted in another tab; point at the tag field
  // and refresh the tag list instead of showing a generic validation toast.
  function handleError(data: unknown): boolean {
    if (!hasDeletedTagError(data)) return false
    form.setError('tagIds', { message: t('validation.tagNotFound') })
    queryClient.invalidateQueries({ queryKey: tagsKeys.all })
    return true
  }

  const onSubmit = async (data: EventFormValues) => {
    const payload = toRequestPayload(data, tagsEditable)
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
      handleError,
    })
  }

  return { form, onSubmit, tagsEditable }
}
