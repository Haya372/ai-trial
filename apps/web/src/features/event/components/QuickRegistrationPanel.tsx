import {
  Button,
  Form,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  Input,
  Popover,
  PopoverContent,
} from '@repo/ui'
import { useRef } from 'react'
import { useTranslation } from 'react-i18next'
import type { EventResponse } from '../../../api/generated'
import { useQuickRegistrationForm } from '../hooks/useQuickRegistrationForm'

interface QuickRegistrationPanelProps {
  open: boolean
  anchor: { x: number; y: number }
  start: Date
  onClose: () => void
  onEditDetail: (event: EventResponse) => void
}

export default function QuickRegistrationPanel({
  open,
  anchor,
  start,
  onClose,
  onEditDetail,
}: QuickRegistrationPanelProps) {
  const { t } = useTranslation('event')
  const { form, phase, onSubmit } = useQuickRegistrationForm(start)
  const titleInputRef = useRef<HTMLInputElement>(null)

  if (!open) return null

  const virtualAnchor = {
    getBoundingClientRect: () => new DOMRect(anchor.x, anchor.y, 0, 0),
  }

  return (
    <Popover
      open={open}
      onOpenChange={(isOpen) => {
        if (!isOpen) onClose()
      }}
    >
      <PopoverContent
        anchor={virtualAnchor}
        side="bottom"
        align="start"
        popupProps={{ initialFocus: titleInputRef }}
      >
        {phase.status === 'input' ? (
          <Form {...form}>
            <form
              onSubmit={form.handleSubmit(onSubmit)}
              className="flex flex-col gap-3"
            >
              <FormField
                control={form.control}
                name="title"
                render={({ field, fieldState }) => (
                  <FormItem>
                    <FormLabel>{t('fields.title')}</FormLabel>
                    <Input
                      id={field.name}
                      placeholder={t('quickRegistration.titlePlaceholder')}
                      state={fieldState.error ? 'error' : 'default'}
                      aria-invalid={!!fieldState.error}
                      {...field}
                      ref={(el) => {
                        field.ref(el)
                        titleInputRef.current = el
                      }}
                    />
                    <FormMessage />
                  </FormItem>
                )}
              />
              <div className="flex justify-end">
                <Button
                  type="submit"
                  variant="primary"
                  size="sm"
                  disabled={form.formState.isSubmitting}
                >
                  {form.formState.isSubmitting
                    ? t('quickRegistration.submitButtonLoading')
                    : t('quickRegistration.submitButton')}
                </Button>
              </div>
            </form>
          </Form>
        ) : (
          <div className="flex flex-col gap-2">
            <p className="text-sm">{t('quickRegistration.successMessage')}</p>
            <Button
              type="button"
              variant="link"
              size="sm"
              onClick={() => onEditDetail(phase.event)}
            >
              {t('quickRegistration.editDetailLink')}
            </Button>
          </div>
        )}
      </PopoverContent>
    </Popover>
  )
}
