import { useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from '@repo/ui'
import { deleteSubscription, subscribeToShare } from '../../../api/generated'
import { getSubscribeErrorMessage, getUnsubscribeErrorMessage } from '../utils'

export type SubscriptionStatus =
  | 'notSubscribed'
  | 'subscribing'
  | 'subscribed'
  | 'unsubscribing'

export function useSubscriptionActions(
  token: string,
  initialIsSubscribed: boolean,
): {
  status: SubscriptionStatus
  errorMessage: string | null
  onSubscribe: () => Promise<void>
  onUnsubscribe: () => Promise<void>
  dismissError: () => void
} {
  const { t } = useTranslation(['share', 'common'])
  const [status, setStatus] = useState<SubscriptionStatus>(
    initialIsSubscribed ? 'subscribed' : 'notSubscribed',
  )
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  // POST 成功時に取得した subscriptionId を保持する ref
  const subscriptionIdRef = useRef<string | null>(null)

  async function onSubscribe(): Promise<void> {
    setStatus('subscribing')
    setErrorMessage(null)
    try {
      const res = await subscribeToShare(token)
      if (res.status === 200 || res.status === 201) {
        subscriptionIdRef.current = res.data.id
        setStatus('subscribed')
        toast.success(t('share:subscription.addSuccessToast'))
      } else {
        const msg = getSubscribeErrorMessage(res.data, t)
        setErrorMessage(msg)
        toast.error(msg)
        setStatus('notSubscribed')
      }
    } catch (err) {
      const msg = getSubscribeErrorMessage(err, t)
      setErrorMessage(msg)
      toast.error(msg)
      setStatus('notSubscribed')
    }
  }

  async function onUnsubscribe(): Promise<void> {
    setStatus('unsubscribing')
    setErrorMessage(null)

    // 初期 isSubscribed=true の場合など id が未取得のとき冪等 POST で id を取得する
    if (subscriptionIdRef.current === null) {
      try {
        const res = await subscribeToShare(token)
        if (res.status === 200 || res.status === 201) {
          subscriptionIdRef.current = res.data.id
        } else {
          const msg = getSubscribeErrorMessage(res.data, t)
          setErrorMessage(msg)
          toast.error(msg)
          setStatus('subscribed')
          return
        }
      } catch (err) {
        const msg = getSubscribeErrorMessage(err, t)
        setErrorMessage(msg)
        toast.error(msg)
        setStatus('subscribed')
        return
      }
    }

    try {
      const id = subscriptionIdRef.current
      const res = await deleteSubscription(id)
      if (res.status === 204) {
        subscriptionIdRef.current = null
        setStatus('notSubscribed')
        toast.success(t('share:subscription.removeSuccessToast'))
      } else {
        const msg = getUnsubscribeErrorMessage(res.data, t)
        setErrorMessage(msg)
        toast.error(msg)
        setStatus('subscribed')
      }
    } catch (err) {
      const msg = getUnsubscribeErrorMessage(err, t)
      setErrorMessage(msg)
      toast.error(msg)
      setStatus('subscribed')
    }
  }

  function dismissError(): void {
    setErrorMessage(null)
  }

  return { status, errorMessage, onSubscribe, onUnsubscribe, dismissError }
}
