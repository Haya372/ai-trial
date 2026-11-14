import { Button, TagBadge } from '@repo/ui'
import { useTranslation } from 'react-i18next'
import type { TagResponse } from '../../../api/generated'

interface TagListItemProps {
  tag: TagResponse
  onEdit: (tag: TagResponse) => void
  onDelete: (tag: TagResponse) => void
}

export default function TagListItem({
  tag,
  onEdit,
  onDelete,
}: TagListItemProps) {
  const { t } = useTranslation('tag')
  return (
    <li className="flex items-center justify-between gap-2 py-1">
      <div className="min-w-0">
        <TagBadge name={tag.name} color={tag.color} />
      </div>
      <div className="flex shrink-0 gap-1">
        <Button
          variant="ghost"
          size="sm"
          aria-label={t('management.edit', { name: tag.name })}
          onClick={() => onEdit(tag)}
        >
          {t('management.editButton')}
        </Button>
        <Button
          variant="destructive"
          size="sm"
          aria-label={t('management.delete', { name: tag.name })}
          onClick={() => onDelete(tag)}
        >
          {t('management.deleteButton')}
        </Button>
      </div>
    </li>
  )
}
