import { zodResolver } from '@hookform/resolvers/zod'
import { TAG_COLOR_KEYS } from '@repo/ui'
import { useQueryClient } from '@tanstack/react-query'
import { useMemo, useState } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { createTag, updateTag } from '../../../api/generated'
import { useRevalidateOnLanguageChange } from '../../../hooks/useRevalidateOnLanguageChange'
import { hasErrorCode } from '../../../lib/errorMessage'
import { eventsKeys, tagsKeys } from '../../../lib/queryKeys'
import {
  createTagFormSchema,
  type TagEditorTarget,
  type TagFormValues,
} from '../types'

function toFormValues(target: TagEditorTarget): TagFormValues {
  return target.mode === 'edit'
    ? { name: target.tag.name, color: target.tag.color }
    : { name: '', color: TAG_COLOR_KEYS[0] }
}

export function useTagEditorForm(
  target: TagEditorTarget,
  onSaved: () => void,
  onLimitExceeded: () => void,
) {
  const { t } = useTranslation('tag')
  const queryClient = useQueryClient()
  const schema = useMemo(() => createTagFormSchema(t), [t])
  const form = useForm<TagFormValues>({
    resolver: zodResolver(schema),
    mode: 'onTouched',
    defaultValues: toFormValues(target),
  })
  const [submitError, setSubmitError] = useState<string | null>(null)

  useRevalidateOnLanguageChange(form)

  // Duplicate names and the tag limit are decided by the server's error code,
  // not by checking the cached list, so concurrent edits elsewhere are respected.
  function handleError(data: unknown) {
    const code = hasErrorCode(data) ? data.code : undefined
    if (code === 'TAG_NAME_TAKEN') {
      form.setError('name', { message: t('errors.nameTaken') })
    } else if (code === 'TAG_LIMIT_EXCEEDED') {
      onLimitExceeded()
    } else {
      setSubmitError(t('errors.saveFailed'))
    }
  }

  const onSubmit = async (values: TagFormValues) => {
    setSubmitError(null)
    try {
      const res =
        target.mode === 'create'
          ? await createTag(values)
          : await updateTag(target.tag.id, values)
      const expectedStatus = target.mode === 'create' ? 201 : 200
      if (res.status !== expectedStatus) {
        handleError(res.data)
        return
      }
      await queryClient.invalidateQueries({ queryKey: tagsKeys.all })
      // Renaming or recoloring changes how tagged events render on the calendar.
      if (target.mode === 'edit') {
        await queryClient.invalidateQueries({ queryKey: eventsKeys.all })
      }
      onSaved()
    } catch {
      setSubmitError(t('errors.saveFailed'))
    }
  }

  return { form, onSubmit, submitError }
}
