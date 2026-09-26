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
  Textarea,
} from '@repo/ui'
import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import type { EventResponse } from '../../../api/generated'
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
  const { t } = useTranslation('event')
  const { form, onSubmit } = useEventForm(
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
            <DialogFooter>
              <Button type="button" variant="secondary" onClick={onClose}>
                {t('form.cancelButton')}
              </Button>
              <Button
                type="submit"
                variant="primary"
                disabled={form.formState.isSubmitting}
              >
                {form.formState.isSubmitting
                  ? t('form.submitButtonLoading')
                  : t('form.submitButton')}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
