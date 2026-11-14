import { Button } from '@repo/ui'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import TagManagementModal from './TagManagementModal'

export default function TagManagementButton() {
  const { t } = useTranslation('tag')
  const [open, setOpen] = useState(false)
  return (
    <>
      <Button variant="outline" onClick={() => setOpen(true)}>
        {t('management.openButton')}
      </Button>
      <TagManagementModal open={open} onClose={() => setOpen(false)} />
    </>
  )
}
