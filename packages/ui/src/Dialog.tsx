'use client'

import { Dialog as DialogPrimitive } from '@base-ui/react/dialog'
import { cn } from 'cn'

// Re-export primitive parts directly
const Dialog = DialogPrimitive.Root
const DialogTrigger = DialogPrimitive.Trigger
const DialogClose = DialogPrimitive.Close

// DialogContent: Portal + Backdrop + Popup wrapper
type DialogContentProps = Omit<DialogPrimitive.Popup.Props, 'className'>

function DialogContent({ children, ...props }: DialogContentProps) {
  return (
    <DialogPrimitive.Portal>
      <DialogPrimitive.Backdrop
        className={cn(
          'fixed inset-0 z-50 bg-black/50',
          'data-[ending-style]:opacity-0 data-[starting-style]:opacity-0',
          'transition-opacity duration-200',
        )}
      />
      <DialogPrimitive.Popup
        data-slot="dialog-content"
        className={cn(
          'fixed left-1/2 top-1/2 z-50 -translate-x-1/2 -translate-y-1/2',
          'w-full max-w-lg',
          'rounded-xl border border-border bg-background shadow-xl',
          'p-6',
          'focus:outline-none',
          'data-[ending-style]:opacity-0 data-[ending-style]:scale-95',
          'data-[starting-style]:opacity-0 data-[starting-style]:scale-95',
          'transition-all duration-200',
        )}
        {...props}
      >
        {children}
      </DialogPrimitive.Popup>
    </DialogPrimitive.Portal>
  )
}

// DialogHeader
type DialogHeaderProps = Omit<React.HTMLAttributes<HTMLDivElement>, 'className'>

function DialogHeader({ ...props }: DialogHeaderProps) {
  return (
    <div
      data-slot="dialog-header"
      className={cn('flex flex-col gap-1.5 pb-4')}
      {...props}
    />
  )
}

// DialogFooter
type DialogFooterProps = Omit<React.HTMLAttributes<HTMLDivElement>, 'className'>

function DialogFooter({ ...props }: DialogFooterProps) {
  return (
    <div
      data-slot="dialog-footer"
      className={cn('flex flex-row justify-end gap-2 pt-4')}
      {...props}
    />
  )
}

// DialogTitle
type DialogTitleProps = Omit<DialogPrimitive.Title.Props, 'className'>

function DialogTitle({ ...props }: DialogTitleProps) {
  return (
    <DialogPrimitive.Title
      data-slot="dialog-title"
      className={cn('text-lg font-semibold leading-none tracking-tight')}
      {...props}
    />
  )
}

// DialogDescription
type DialogDescriptionProps = Omit<
  DialogPrimitive.Description.Props,
  'className'
>

function DialogDescription({ ...props }: DialogDescriptionProps) {
  return (
    <DialogPrimitive.Description
      data-slot="dialog-description"
      className={cn('text-sm text-muted-foreground')}
      {...props}
    />
  )
}

export {
  Dialog,
  DialogTrigger,
  DialogContent,
  DialogHeader,
  DialogFooter,
  DialogTitle,
  DialogDescription,
  DialogClose,
}
