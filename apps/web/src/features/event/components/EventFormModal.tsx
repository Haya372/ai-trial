import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Form,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  Input,
  Label,
  Textarea,
} from '@repo/ui'
import { useEffect, useId } from 'react'
import { useTranslation } from 'react-i18next'
import type { EventResponse } from '../../../api/generated'
import EventTagSelector from '../../tag/components/EventTagSelector'
import { useEventForm } from '../hooks/useEventForm'
import type { EventFormMode } from '../types'

interface EventFormModalProps {
  open: boolean
  mode: EventFormMode
  event: EventResponse | null
  initialStart?: Date | null
  onClose: () => void
}

export default function EventFormModal({
  open,
  mode,
  event,
  initialStart,
  onClose,
}: EventFormModalProps) {
  const { t } = useTranslation(['event', 'tag'])
  const tagLabelId = useId()
  const { form, onSubmit, tagsEditable } = useEventForm(
    open,
    mode,
    event,
    initialStart,
    onClose,
  )

  useEffect(() => {
    if (open) form.setFocus('title')
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  return (
    <Dialog
      open={open}
      onOpenChange={(isOpen) => {
        if (!isOpen) onClose()
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {mode === 'create' ? t('form.createTitle') : t('form.editTitle')}
          </DialogTitle>
        </DialogHeader>
        <Form {...form}>
          <form
            onSubmit={form.handleSubmit(onSubmit)}
            className="flex flex-col gap-4"
          >
            <FormField
              control={form.control}
              name="title"
              render={({ field, fieldState }) => (
                <FormItem>
                  <FormLabel>{t('fields.title')}</FormLabel>
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
              name="startAt"
              render={({ field, fieldState }) => (
                <FormItem>
                  <FormLabel>{t('fields.startAt')}</FormLabel>
                  <Input
                    id={field.name}
                    type="datetime-local"
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
              name="endAt"
              render={({ field, fieldState }) => (
                <FormItem>
                  <FormLabel>{t('fields.endAt')}</FormLabel>
                  <Input
                    id={field.name}
                    type="datetime-local"
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
              name="description"
              render={({ field, fieldState }) => (
                <FormItem>
                  <FormLabel>{t('fields.description')}</FormLabel>
                  <Textarea
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
              name="location"
              render={({ field, fieldState }) => (
                <FormItem>
                  <FormLabel>{t('fields.location')}</FormLabel>
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
              name="url"
              render={({ field, fieldState }) => (
                <FormItem>
                  <FormLabel>{t('fields.url')}</FormLabel>
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
            {tagsEditable && (
              <FormField
                control={form.control}
                name="tagIds"
                render={({ field, fieldState }) => (
                  <FormItem>
                    <Label id={tagLabelId}>{t('tag:selector.label')}</Label>
                    <EventTagSelector
                      labelId={tagLabelId}
                      value={field.value}
                      onChange={field.onChange}
                      initialTags={event?.tags ?? []}
                      invalid={!!fieldState.error}
                    />
                    <FormMessage />
                  </FormItem>
                )}
              />
            )}
            <DialogFooter>
              <Button type="button" variant="secondary" onClick={onClose}>
                {t('form.cancelButton')}
              </Button>
              <Button
                type="submit"
                variant="primary"
                loading={form.formState.isSubmitting}
              >
                {t('form.submitButton')}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
