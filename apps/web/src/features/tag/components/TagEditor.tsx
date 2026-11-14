import {
  Alert,
  AlertDescription,
  Button,
  Form,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  Input,
  Label,
  TAG_COLOR_KEYS,
  type TagColorKey,
  TagColorPalette,
} from '@repo/ui'
import { useEffect, useId, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { useTagEditorForm } from '../hooks/useTagEditorForm'
import type { TagEditorTarget } from '../types'

interface TagEditorProps {
  target: TagEditorTarget
  onDone: () => void
  onLimitExceeded: () => void
}

export default function TagEditor({
  target,
  onDone,
  onLimitExceeded,
}: TagEditorProps) {
  const { t } = useTranslation('tag')
  const colorLabelId = useId()
  const { form, onSubmit, submitError } = useTagEditorForm(
    target,
    onDone,
    onLimitExceeded,
  )
  const colorLabels = useMemo(
    () =>
      Object.fromEntries(
        TAG_COLOR_KEYS.map((color) => [color, t(`colors.${color}`)]),
      ) as Record<TagColorKey, string>,
    [t],
  )

  useEffect(() => {
    form.setFocus('name')
  }, [form])

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className="flex flex-col gap-3 rounded-md border border-border p-3"
      >
        {submitError && (
          <Alert variant="destructive">
            <AlertDescription>{submitError}</AlertDescription>
          </Alert>
        )}
        <FormField
          control={form.control}
          name="name"
          render={({ field, fieldState }) => (
            <FormItem>
              <FormLabel>{t('editor.name')}</FormLabel>
              <Input
                id={field.name}
                state={fieldState.error ? 'error' : 'default'}
                aria-invalid={!!fieldState.error}
                {...field}
              />
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="color"
          render={({ field }) => (
            <FormItem>
              <Label id={colorLabelId}>{t('editor.color')}</Label>
              <TagColorPalette
                value={field.value}
                onChange={field.onChange}
                colorLabels={colorLabels}
                aria-labelledby={colorLabelId}
              />
            </FormItem>
          )}
        />
        <div className="flex justify-end gap-2">
          <Button type="button" variant="secondary" onClick={onDone}>
            {t('editor.cancel')}
          </Button>
          <Button type="submit" loading={form.formState.isSubmitting}>
            {t('editor.save')}
          </Button>
        </div>
      </form>
    </Form>
  )
}
