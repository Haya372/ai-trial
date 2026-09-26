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
            {mode === 'create' ? '予定を作成' : '予定を編集'}
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
                  <FormLabel>タイトル</FormLabel>
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
                  <FormLabel>開始日時</FormLabel>
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
                  <FormLabel>終了日時</FormLabel>
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
                  <FormLabel>メモ</FormLabel>
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
                  <FormLabel>場所</FormLabel>
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
                  <FormLabel>URL</FormLabel>
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
                キャンセル
              </Button>
              <Button
                type="submit"
                variant="primary"
                disabled={form.formState.isSubmitting}
              >
                {form.formState.isSubmitting ? '保存中…' : '保存'}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
