import {
  Alert,
  AlertDescription,
  Button,
  Loading,
  Popover,
  PopoverContent,
  PopoverTrigger,
  TagBadge,
} from '@repo/ui'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { TagResponse } from '../../../api/generated'
import { useTagsQuery } from '../hooks/useTagsQuery'
import TagManagementModal from './TagManagementModal'

interface TagFilterButtonProps {
  selectedTagId: string | null
  onChange: (tagId: string | null) => void
}

export default function TagFilterButton({
  selectedTagId,
  onChange,
}: TagFilterButtonProps) {
  const { t } = useTranslation('tag')
  const [open, setOpen] = useState(false)
  const [managementOpen, setManagementOpen] = useState(false)
  // While filtering, the trigger shows the selected tag's name and color, so the
  // list stays subscribed even with the popover closed.
  const { data: tags, isError } = useTagsQuery(open || selectedTagId !== null)
  const selectedTag = tags?.find((tag) => tag.id === selectedTagId)

  function select(tagId: string | null) {
    onChange(tagId)
    setOpen(false)
  }

  function openManagement() {
    setOpen(false)
    setManagementOpen(true)
  }

  function renderList(tagList: TagResponse[]) {
    if (tagList.length === 0) {
      return (
        <div className="flex flex-col items-center gap-2 py-2">
          <p className="text-sm text-muted-foreground">{t('filter.empty')}</p>
          <Button variant="outline" size="sm" onClick={openManagement}>
            {t('filter.openManagement')}
          </Button>
        </div>
      )
    }
    return (
      <div className="flex flex-col gap-1">
        <div className="flex">
          <Button
            variant="ghost"
            size="sm"
            aria-pressed={selectedTagId === null}
            onClick={() => select(null)}
          >
            {t('filter.showAll')}
          </Button>
        </div>
        <ul className="flex max-h-60 flex-col overflow-y-auto border-t border-border pt-1">
          {tagList.map((tag) => (
            <li key={tag.id} className="flex">
              <Button
                variant="ghost"
                size="sm"
                aria-pressed={tag.id === selectedTagId}
                onClick={() => select(tag.id)}
              >
                <TagBadge name={tag.name} color={tag.color} size="sm" />
              </Button>
            </li>
          ))}
        </ul>
      </div>
    )
  }

  function renderContent() {
    // A failed background refetch keeps the last list, which is still pickable.
    if (tags) return renderList(tags)
    if (isError) {
      return (
        <Alert variant="destructive">
          <AlertDescription>{t('filter.loadError')}</AlertDescription>
        </Alert>
      )
    }
    return <Loading label={t('filter.loading')} />
  }

  return (
    <>
      <div className="flex items-center">
        <Popover open={open} onOpenChange={setOpen}>
          <PopoverTrigger
            render={
              <Button
                variant="outline"
                aria-label={
                  selectedTag
                    ? t('filter.active', { name: selectedTag.name })
                    : undefined
                }
              />
            }
          >
            {selectedTag ? (
              <TagBadge
                name={selectedTag.name}
                color={selectedTag.color}
                size="sm"
              />
            ) : (
              t('filter.button')
            )}
          </PopoverTrigger>
          <PopoverContent align="end">{renderContent()}</PopoverContent>
        </Popover>
        {selectedTagId !== null && (
          <Button
            variant="ghost"
            size="sm"
            aria-label={t('filter.clear')}
            onClick={() => onChange(null)}
          >
            ×
          </Button>
        )}
      </div>
      <TagManagementModal
        open={managementOpen}
        onClose={() => setManagementOpen(false)}
      />
    </>
  )
}
