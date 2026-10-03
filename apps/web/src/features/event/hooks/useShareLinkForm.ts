import { zodResolver } from '@hookform/resolvers/zod'
import { toast } from '@repo/ui'
import { useEffect, useMemo, useRef, useState } from 'react'
import type { SubmitHandler, UseFormReturn } from 'react-hook-form'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import type {
  EventResponse,
  ValidationErrorResponse,
} from '../../../api/generated'
import { createEventShare } from '../../../api/generated'
import { useRevalidateOnLanguageChange } from '../../../hooks/useRevalidateOnLanguageChange'
import { createShareLinkFormSchema, type ShareLinkFormValues } from '../types'
import {
  defaultExpiresAtValue,
  getShareErrorMessage,
  toFullShareUrl,
  toIsoString,
} from '../utils'

export type ShareLinkResult = {
  shareUrl: string
  expiresAt: string
}

function findExpiresAtDetail(
  data: unknown,
): { field: 'expiresAt'; code: string } | null {
  if (
    typeof data !== 'object' ||
    data === null ||
    !('details' in data) ||
    !Array.isArray((data as { details: unknown }).details)
  ) {
    return null
  }
  const details = (data as ValidationErrorResponse).details
  const detail = details.find(
    (d) => d.field === 'expiresAt' && typeof d.code === 'string',
  )
  return detail ? { field: 'expiresAt', code: detail.code } : null
}

function defaultFormValues(event: EventResponse | null): ShareLinkFormValues {
  return {
    expiresAt: event ? defaultExpiresAtValue(event) : '',
  }
}

export function useShareLinkForm(
  open: boolean,
  event: EventResponse | null,
  onClose: () => void,
): {
  form: UseFormReturn<ShareLinkFormValues>
  result: ShareLinkResult | null
  errorMessage: string | null
  onSubmit: SubmitHandler<ShareLinkFormValues>
  resetAll: () => void
  resetResult: () => void
} {
  const { t } = useTranslation(['eventshare', 'common', 'event'])
  const schema = useMemo(() => createShareLinkFormSchema(t, event), [t, event])
  const form = useForm<ShareLinkFormValues>({
    resolver: zodResolver(schema),
    mode: 'onTouched',
    defaultValues: defaultFormValues(event),
  })

  const [result, setResult] = useState<ShareLinkResult | null>(null)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)

  // onSubmit は呼び出し時点の event をクロージャで保持するため、生成中に
  // ダイアログが別の予定に切り替わった後も古いレスポンスで状態更新してしまう。
  // 現在表示中の予定IDをレンダーごとに追跡し、レスポンス到達時に不一致なら無視する。
  const currentEventIdRef = useRef<string | null>(null)
  currentEventIdRef.current = event?.id ?? null

  useEffect(() => {
    if (open) {
      form.reset(defaultFormValues(event))
      setResult(null)
      setErrorMessage(null)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  useRevalidateOnLanguageChange(form)

  function resetAll() {
    form.reset(defaultFormValues(event))
    setResult(null)
    setErrorMessage(null)
    onClose()
  }

  const onSubmit: SubmitHandler<ShareLinkFormValues> = async (data) => {
    if (!event) return
    const requestedEventId = event.id

    try {
      const res = await createEventShare(event.id, {
        expiresAt: toIsoString(data.expiresAt),
      })

      // ダイアログが別の予定に切り替わった後に届いた古いレスポンスは無視する
      if (currentEventIdRef.current !== requestedEventId) return

      if (res.status === 201) {
        setResult({
          shareUrl: toFullShareUrl(res.data.url),
          expiresAt: res.data.expiresAt,
        })
        toast.success(t('eventshare:toast.createSuccess'))
        return
      }

      const expiresAtDetail =
        res.status === 400 ? findExpiresAtDetail(res.data) : null
      if (expiresAtDetail) {
        const code = expiresAtDetail.code
        const keyMap: Record<string, string> = {
          BEFORE_EVENT_START: t('eventshare:validation.expiresAtAfterStart'),
          NOT_IN_FUTURE: t('eventshare:validation.expiresAtInFuture'),
        }
        const message =
          keyMap[code] ?? t('eventshare:errors.validationFallback')
        form.setError('expiresAt', { message })
        return
      }

      const message = getShareErrorMessage(res.data, t)
      setErrorMessage(message)
      toast.error(message)
    } catch {
      if (currentEventIdRef.current !== requestedEventId) return
      const message = t('eventshare:errors.createFallback')
      setErrorMessage(message)
      toast.error(message)
    }
  }

  function resetResult() {
    setResult(null)
    setErrorMessage(null)
  }

  return { form, result, errorMessage, onSubmit, resetAll, resetResult }
}
