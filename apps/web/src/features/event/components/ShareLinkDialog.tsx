import {
  Alert,
  AlertDescription,
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
  Text,
  toast,
} from '@repo/ui'
import { useTranslation } from 'react-i18next'
import type { EventResponse } from '../../../api/generated'
import { useShareLinkForm } from '../hooks/useShareLinkForm'

interface ShareLinkDialogProps {
  open: boolean
  event: EventResponse | null
  onClose: () => void
}

export default function ShareLinkDialog({
  open,
  event,
  onClose,
}: ShareLinkDialogProps) {
  const { t } = useTranslation('eventshare')
  const { form, result, errorMessage, onSubmit, resetAll, resetResult } =
    useShareLinkForm(open, event, onClose)

  async function copyShareUrl(url: string): Promise<void> {
    try {
      await navigator.clipboard.writeText(url)
      toast.success(t('toast.copySuccess'))
    } catch {
      toast.error(t('toast.copyFailure'))
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(isOpen) => {
        if (!isOpen) resetAll()
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t('dialog.title')}</DialogTitle>
        </DialogHeader>

        <Text variant="caption">{t('dialog.descriptionWarning')}</Text>

        {errorMessage && (
          <Alert role="alert" variant="destructive">
            <AlertDescription>{errorMessage}</AlertDescription>
          </Alert>
        )}

        {result ? (
          <>
            <div className="flex flex-col gap-2">
              <FormLabel>{t('dialog.shareUrlLabel')}</FormLabel>
              <div className="flex gap-2">
                <Input
                  value={result.shareUrl}
                  readOnly
                  onFocus={(e) => e.currentTarget.select()}
                />
                <Button
                  type="button"
                  variant="secondary"
                  onClick={() => copyShareUrl(result.shareUrl)}
                >
                  {t('dialog.copy')}
                </Button>
              </div>
              <Text variant="caption">
                {t('dialog.expiresAtInfo', {
                  dateTime: new Date(result.expiresAt).toLocaleString(),
                })}
              </Text>
            </div>
            <DialogFooter>
              <Button type="button" variant="secondary" onClick={resetResult}>
                {t('dialog.regenerate')}
              </Button>
              <Button type="button" variant="primary" onClick={resetAll}>
                {t('dialog.close')}
              </Button>
            </DialogFooter>
          </>
        ) : (
          <Form {...form}>
            <form
              onSubmit={form.handleSubmit(onSubmit)}
              className="flex flex-col gap-4"
            >
              <FormField
                control={form.control}
                name="expiresAt"
                render={({ field, fieldState }) => (
                  <FormItem>
                    <FormLabel>{t('dialog.expiresAtLabel')}</FormLabel>
                    <Input
                      id={field.name}
                      type="datetime-local"
                      state={fieldState.error ? 'error' : 'default'}
                      aria-invalid={!!fieldState.error}
                      aria-label={t('dialog.expiresAtLabel')}
                      {...field}
                    />
                    <FormMessage />
                  </FormItem>
                )}
              />
              <DialogFooter>
                <Button type="button" variant="secondary" onClick={resetAll}>
                  {t('dialog.cancel')}
                </Button>
                <Button
                  type="submit"
                  variant="primary"
                  disabled={form.formState.isSubmitting}
                >
                  {form.formState.isSubmitting
                    ? t('dialog.generating')
                    : t('dialog.generate')}
                </Button>
              </DialogFooter>
            </form>
          </Form>
        )}
      </DialogContent>
    </Dialog>
  )
}
