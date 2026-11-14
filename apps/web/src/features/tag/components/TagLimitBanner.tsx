import { Alert, AlertDescription } from '@repo/ui'
import { useTranslation } from 'react-i18next'
import { TAG_LIMIT } from '../types'

export default function TagLimitBanner() {
  const { t } = useTranslation('tag')
  return (
    <Alert variant="warning">
      <AlertDescription>
        {t('management.limitReached', { limit: TAG_LIMIT })}
      </AlertDescription>
    </Alert>
  )
}
