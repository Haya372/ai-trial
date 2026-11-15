import {
  Alert,
  AlertDescription,
  Button,
  Checkbox,
  Loading,
  Popover,
  PopoverContent,
  PopoverTrigger,
  TagBadge,
} from '@repo/ui'
import { useId, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { EventTagResponse, TagResponse } from '../../../api/generated'
import { useTagsQuery } from '../hooks/useTagsQuery'
import { EVENT_TAG_LIMIT, TAG_LIMIT } from '../types'
import TagEditor from './TagEditor'
import TagLimitBanner from './TagLimitBanner'

interface EventTagSelectorProps {
  labelId: string
  value: string[]
  onChange: (tagIds: string[]) => void
  initialTags: EventTagResponse[]
  invalid?: boolean
}

export default function EventTagSelector({
  labelId,
  value,
  onChange,
  initialTags,
  invalid,
}: EventTagSelectorProps) {
  const { t } = useTranslation('tag')
  const idPrefix = useId()
  const { data: tags, isPending, isError } = useTagsQuery(true)
  const [creating, setCreating] = useState(false)
  // The server can reject a create with TAG_LIMIT_EXCEEDED even when the cached
  // list is below the limit (e.g. tags were added in another tab).
  const [limitExceededByServer, setLimitExceededByServer] = useState(false)

  // initialTags keeps tags deleted elsewhere visible so the user can remove them.
  const tagById = useMemo(
    () =>
      new Map<string, EventTagResponse>(
        [...initialTags, ...(tags ?? [])].map((tag) => [tag.id, tag]),
      ),
    [initialTags, tags],
  )
  const selectedTags = value.flatMap((id) => {
    const tag = tagById.get(id)
    return tag ? [tag] : []
  })
  const selectionFull = value.length >= EVENT_TAG_LIMIT
  const tagLimitReached =
    limitExceededByServer || (tags?.length ?? 0) >= TAG_LIMIT

  function remove(tagId: string) {
    onChange(value.filter((id) => id !== tagId))
  }

  function toggle(tagId: string, checked: boolean) {
    if (checked) onChange([...value, tagId])
    else remove(tagId)
  }

  function handleCreated(tag: TagResponse) {
    onChange([...value, tag.id])
    setCreating(false)
  }

  function handleLimitExceeded() {
    setLimitExceededByServer(true)
    setCreating(false)
  }

  function renderList(tagList: TagResponse[]) {
    if (tagList.length === 0) {
      return (
        <p className="py-2 text-center text-sm text-muted-foreground">
          {t('selector.empty')}
        </p>
      )
    }
    return (
      <ul className="flex max-h-60 flex-col overflow-y-auto">
        {tagList.map((tag) => {
          const checked = value.includes(tag.id)
          const nameId = `${idPrefix}-${tag.id}`
          return (
            <li key={tag.id}>
              <label className="flex cursor-pointer items-center gap-2 py-1">
                <Checkbox
                  checked={checked}
                  disabled={!checked && selectionFull}
                  onCheckedChange={(next) => toggle(tag.id, next)}
                  aria-labelledby={nameId}
                />
                <span id={nameId} className="min-w-0">
                  <TagBadge name={tag.name} color={tag.color} size="sm" />
                </span>
              </label>
            </li>
          )
        })}
      </ul>
    )
  }

  function renderContent() {
    if (isPending) return <Loading label={t('selector.loading')} />
    if (isError) {
      return (
        <Alert variant="destructive">
          <AlertDescription>{t('selector.loadError')}</AlertDescription>
        </Alert>
      )
    }
    if (creating) {
      return (
        <TagEditor
          target={{ mode: 'create' }}
          onSaved={handleCreated}
          onCancel={() => setCreating(false)}
          onLimitExceeded={handleLimitExceeded}
        />
      )
    }
    return (
      <div className="flex flex-col gap-2">
        {selectionFull && (
          <Alert variant="warning">
            <AlertDescription>
              {t('selector.limitReached', { limit: EVENT_TAG_LIMIT })}
            </AlertDescription>
          </Alert>
        )}
        {tagLimitReached && <TagLimitBanner />}
        {renderList(tags)}
        <Button
          type="button"
          variant="outline"
          size="sm"
          disabled={selectionFull || tagLimitReached}
          onClick={() => setCreating(true)}
        >
          {t('selector.createTag')}
        </Button>
      </div>
    )
  }

  return (
    <div
      role="group"
      aria-labelledby={labelId}
      className="flex flex-wrap items-center gap-1"
    >
      {selectedTags.map((tag) => (
        <span key={tag.id} className="inline-flex min-w-0 items-center">
          <TagBadge name={tag.name} color={tag.color} size="sm" />
          <Button
            type="button"
            variant="ghost"
            size="sm"
            aria-label={t('selector.remove', { name: tag.name })}
            onClick={() => remove(tag.id)}
          >
            ×
          </Button>
        </span>
      ))}
      <Popover
        onOpenChange={(open) => {
          if (!open) setCreating(false)
        }}
      >
        <PopoverTrigger
          render={
            <Button
              type="button"
              variant="outline"
              size="sm"
              aria-invalid={invalid || undefined}
            />
          }
        >
          {t('selector.open')}
        </PopoverTrigger>
        <PopoverContent align="start">{renderContent()}</PopoverContent>
      </Popover>
    </div>
  )
}
