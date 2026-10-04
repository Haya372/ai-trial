import { useTranslation } from 'react-i18next'
import type { ShareDetailError } from '../hooks/useShareDetailQuery'
import { useShareDetailQuery } from '../hooks/useShareDetailQuery'
import { useViewerAuthQuery } from '../hooks/useViewerAuthQuery'
import ShareEventDetail from '../components/ShareEventDetail'
import SubscriptionActions from '../components/SubscriptionActions'

interface SharePageProps {
  token: string
}

function ErrorMessage({
  title,
  description,
}: {
  title: string
  description: string
}) {
  return (
    <div className="text-center py-8">
      <p className="text-lg font-semibold">{title}</p>
      <p className="text-sm text-muted-foreground mt-1">{description}</p>
    </div>
  )
}

function resolveError(error: unknown): ShareDetailError {
  if (error !== null && typeof error === 'object' && 'kind' in error) {
    return error as ShareDetailError
  }
  return { kind: 'unknown' }
}

export default function SharePage({ token }: SharePageProps) {
  const { t } = useTranslation('share')
  const shareDetail = useShareDetailQuery(token)
  const viewerAuth = useViewerAuthQuery()

  if (shareDetail.isPending) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <p>{t('page.loading')}</p>
      </div>
    )
  }

  if (shareDetail.isError) {
    const err = resolveError(shareDetail.error)

    if (err.kind === 'expired') {
      return (
        <main className="max-w-lg mx-auto px-4 py-8">
          <h1 className="text-lg font-bold mb-4">{t('page.appName')}</h1>
          <ErrorMessage
            title={t('page.expiredTitle')}
            description={t('page.expiredDescription')}
          />
        </main>
      )
    }

    if (err.kind === 'notFound') {
      return (
        <main className="max-w-lg mx-auto px-4 py-8">
          <h1 className="text-lg font-bold mb-4">{t('page.appName')}</h1>
          <ErrorMessage
            title={t('page.notFoundTitle')}
            description={t('page.notFoundDescription')}
          />
        </main>
      )
    }

    return (
      <main className="max-w-lg mx-auto px-4 py-8">
        <h1 className="text-lg font-bold mb-4">{t('page.appName')}</h1>
        <ErrorMessage
          title={t('page.serverErrorTitle')}
          description={t('page.serverErrorDescription')}
        />
      </main>
    )
  }

  const event = shareDetail.data
  // 認証判定が完了するまで SubscriptionActions を表示しない（ちらつき防止）
  const showSubscriptionActions =
    !viewerAuth.isPending && viewerAuth.data !== null && !event.isOwnEvent

  return (
    <main className="max-w-lg mx-auto px-4 py-8">
      <h1 className="text-lg font-bold mb-6">{t('page.appName')}</h1>
      <ShareEventDetail event={event} />
      {showSubscriptionActions && (
        <div className="mt-6 border-t pt-4">
          <SubscriptionActions
            token={token}
            initialIsSubscribed={event.isSubscribed}
          />
        </div>
      )}
    </main>
  )
}
