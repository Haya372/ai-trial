import {
  Alert,
  AlertDescription,
  Button,
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@repo/ui'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { TagResponse } from '../../../api/generated'
import { useTagDelete } from '../hooks/useTagDelete'
import { useTagsQuery } from '../hooks/useTagsQuery'
import { TAG_LIMIT, type TagEditorTarget } from '../types'
import TagDeleteConfirmDialog from './TagDeleteConfirmDialog'
import TagEditor from './TagEditor'
import TagLimitBanner from './TagLimitBanner'
import TagList from './TagList'

interface TagManagementModalProps {
  open: boolean
  onClose: () => void
}

export default function TagManagementModal({
  open,
  onClose,
}: TagManagementModalProps) {
  const { t } = useTranslation('tag')
  const { data: tags, isPending, isError } = useTagsQuery(open)
  const [editorTarget, setEditorTarget] = useState<TagEditorTarget | null>(null)
  // The server can reject a create with TAG_LIMIT_EXCEEDED even when the cached
  // list is below the limit (e.g. tags were added in another tab).
  const [limitExceededByServer, setLimitExceededByServer] = useState(false)
  const [deleteFailed, setDeleteFailed] = useState(false)
  const tagDelete = useTagDelete(
    () => setLimitExceededByServer(false),
    () => setDeleteFailed(true),
  )

  const limitReached = limitExceededByServer || (tags?.length ?? 0) >= TAG_LIMIT

  function handleClose() {
    setEditorTarget(null)
    setLimitExceededByServer(false)
    setDeleteFailed(false)
    onClose()
  }

  function openEditor(target: TagEditorTarget) {
    setDeleteFailed(false)
    setEditorTarget(target)
  }

  function handleLimitExceeded() {
    setLimitExceededByServer(true)
    setEditorTarget(null)
  }

  function requestDelete(tag: TagResponse) {
    setDeleteFailed(false)
    tagDelete.requestDelete(tag)
  }

  function renderEditor(target: TagEditorTarget) {
    return (
      <TagEditor
        target={target}
        onDone={() => setEditorTarget(null)}
        onLimitExceeded={handleLimitExceeded}
      />
    )
  }

  return (
    <>
      <Dialog
        open={open}
        onOpenChange={(isOpen) => {
          if (!isOpen) handleClose()
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('management.title')}</DialogTitle>
          </DialogHeader>
          <div className="flex flex-col gap-3">
            {isPending && (
              <p className="py-4 text-center text-sm text-muted-foreground">
                {t('management.loading')}
              </p>
            )}
            {isError && (
              <Alert variant="destructive">
                <AlertDescription>{t('management.loadError')}</AlertDescription>
              </Alert>
            )}
            {tags && (
              <>
                {deleteFailed && (
                  <Alert variant="destructive">
                    <AlertDescription>
                      {t('errors.deleteFailed')}
                    </AlertDescription>
                  </Alert>
                )}
                {limitReached && <TagLimitBanner />}
                <div>
                  <Button
                    variant="outline"
                    disabled={limitReached}
                    onClick={() => openEditor({ mode: 'create' })}
                  >
                    {t('management.newTag')}
                  </Button>
                </div>
                {editorTarget?.mode === 'create' && renderEditor(editorTarget)}
                <TagList
                  tags={tags}
                  editingTagId={
                    editorTarget?.mode === 'edit' ? editorTarget.tag.id : null
                  }
                  renderEditor={(tag) => renderEditor({ mode: 'edit', tag })}
                  onEdit={(tag) => openEditor({ mode: 'edit', tag })}
                  onDelete={requestDelete}
                />
              </>
            )}
          </div>
        </DialogContent>
      </Dialog>

      <TagDeleteConfirmDialog
        tag={tagDelete.target}
        isDeleting={tagDelete.isDeleting}
        onConfirm={tagDelete.confirmDelete}
        onCancel={tagDelete.cancelDelete}
      />
    </>
  )
}
