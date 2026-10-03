import { Button, Text } from '@repo/ui'
import { useTranslation } from 'react-i18next'

interface ErrorFallbackProps {
  error: unknown
  reset: () => void
}

export default function ErrorFallback({ error, reset }: ErrorFallbackProps) {
  const { t } = useTranslation()

  // 画面には出さないが、原因調査のためにコンソールへ残す
  console.error(error)

  return (
    <div
      role="alert"
      className="flex h-screen flex-col items-center justify-center gap-4 p-8 text-center"
    >
      <Text variant="h2">{t('errors.unexpectedTitle')}</Text>
      <Text variant="body">{t('errors.unexpectedDescription')}</Text>
      <Button variant="primary" onClick={reset}>
        {t('errors.retry')}
      </Button>
    </div>
  )
}
