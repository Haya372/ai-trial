import type { ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import type { TagResponse } from '../../../api/generated'
import TagListItem from './TagListItem'

interface TagListProps {
  tags: TagResponse[]
  editingTagId: string | null
  renderEditor: (tag: TagResponse) => ReactNode
  onEdit: (tag: TagResponse) => void
  onDelete: (tag: TagResponse) => void
}

export default function TagList({
  tags,
  editingTagId,
  renderEditor,
  onEdit,
  onDelete,
}: TagListProps) {
  const { t } = useTranslation('tag')

  if (tags.length === 0) {
    return (
      <p className="py-4 text-center text-sm text-muted-foreground">
        {t('management.empty')}
      </p>
    )
  }

  return (
    <ul className="flex max-h-[50vh] flex-col overflow-y-auto">
      {tags.map((tag) =>
        tag.id === editingTagId ? (
          <li key={tag.id} className="py-1">
            {renderEditor(tag)}
          </li>
        ) : (
          <TagListItem
            key={tag.id}
            tag={tag}
            onEdit={onEdit}
            onDelete={onDelete}
          />
        ),
      )}
    </ul>
  )
}
