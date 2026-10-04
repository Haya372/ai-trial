import { Alert, AlertDescription, Button } from '@repo/ui'
import { useTranslation } from 'react-i18next'
import { useSubscriptionActions } from '../hooks/useSubscriptionActions'

interface SubscriptionActionsProps {
  token: string
  initialIsSubscribed: boolean
}

export default function SubscriptionActions({
  token,
  initialIsSubscribed,
}: SubscriptionActionsProps) {
  const { t } = useTranslation('share')
  const { status, errorMessage, onSubscribe, onUnsubscribe, dismissError } =
    useSubscriptionActions(token, initialIsSubscribed)

  return (
    <div className="flex flex-col gap-2">
      {errorMessage != null && (
        <Alert role="alert" variant="destructive">
          <AlertDescription>
            <span>{errorMessage}</span>
            <button
              type="button"
              onClick={dismissError}
              className="ml-2 underline text-sm"
            >
              閉じる
            </button>
          </AlertDescription>
        </Alert>
      )}

      {(status === 'notSubscribed' || status === 'subscribing') && (
        <Button
          onClick={() => void onSubscribe()}
          disabled={status === 'subscribing'}
        >
          {status === 'subscribing'
            ? t('subscription.adding')
            : t('subscription.addButton')}
        </Button>
      )}

      {(status === 'subscribed' || status === 'unsubscribing') && (
        <div className="flex flex-col gap-1">
          <span className="text-sm text-muted-foreground">
            {t('subscription.subscribedLabel')}
          </span>
          <Button
            variant="secondary"
            onClick={() => void onUnsubscribe()}
            disabled={status === 'unsubscribing'}
          >
            {status === 'unsubscribing'
              ? t('subscription.removing')
              : t('subscription.removeButton')}
          </Button>
        </div>
      )}
    </div>
  )
}
