import { Button, Text } from '@repo/ui'
import { useTranslation } from 'react-i18next'

interface ErrorFallbackProps {
  reset: () => void
}

export default function ErrorFallback({ reset }: ErrorFallbackProps) {
  const { t } = useTranslation()

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
