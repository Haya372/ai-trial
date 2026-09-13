'use client'

import { AlertDialog as AlertDialogPrimitive } from '@base-ui/react/alert-dialog'
import { cn } from 'cn'

// Re-export primitive parts directly
const AlertDialog = AlertDialogPrimitive.Root
const AlertDialogTrigger = AlertDialogPrimitive.Trigger

// AlertDialogContent: Portal + Backdrop + Popup wrapper
type AlertDialogContentProps = Omit<
  AlertDialogPrimitive.Popup.Props,
  'className'
>

function AlertDialogContent({ children, ...props }: AlertDialogContentProps) {
  return (
    <AlertDialogPrimitive.Portal>
      <AlertDialogPrimitive.Backdrop
        className={cn(
          'fixed inset-0 z-50 bg-black/50',
          'data-[ending-style]:opacity-0 data-[starting-style]:opacity-0',
          'transition-opacity duration-200',
        )}
      />
      <AlertDialogPrimitive.Popup
        data-slot="alert-dialog-content"
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
      </AlertDialogPrimitive.Popup>
    </AlertDialogPrimitive.Portal>
  )
}

// AlertDialogHeader
type AlertDialogHeaderProps = Omit<
  React.HTMLAttributes<HTMLDivElement>,
  'className'
>

function AlertDialogHeader({ ...props }: AlertDialogHeaderProps) {
  return (
    <div
      data-slot="alert-dialog-header"
      className={cn('flex flex-col gap-1.5 pb-4')}
      {...props}
    />
  )
}

// AlertDialogFooter
type AlertDialogFooterProps = Omit<
  React.HTMLAttributes<HTMLDivElement>,
  'className'
>

function AlertDialogFooter({ ...props }: AlertDialogFooterProps) {
  return (
    <div
      data-slot="alert-dialog-footer"
      className={cn('flex flex-row justify-end gap-2 pt-4')}
      {...props}
    />
  )
}

// AlertDialogTitle
type AlertDialogTitleProps = Omit<AlertDialogPrimitive.Title.Props, 'className'>

function AlertDialogTitle({ ...props }: AlertDialogTitleProps) {
  return (
    <AlertDialogPrimitive.Title
      data-slot="alert-dialog-title"
      className={cn('text-lg font-semibold leading-none tracking-tight')}
      {...props}
    />
  )
}

// AlertDialogDescription
type AlertDialogDescriptionProps = Omit<
  AlertDialogPrimitive.Description.Props,
  'className'
>

function AlertDialogDescription({ ...props }: AlertDialogDescriptionProps) {
  return (
    <AlertDialogPrimitive.Description
      data-slot="alert-dialog-description"
      className={cn('text-sm text-muted-foreground')}
      {...props}
    />
  )
}

// AlertDialogAction — destructive confirm button
type AlertDialogActionProps = Omit<
  AlertDialogPrimitive.Close.Props,
  'className'
>

function AlertDialogAction({ ...props }: AlertDialogActionProps) {
  return (
    <AlertDialogPrimitive.Close
      data-slot="alert-dialog-action"
      className={cn(
        'inline-flex items-center justify-center rounded-lg px-4 py-2 text-sm font-medium',
        'bg-destructive text-white',
        'hover:bg-destructive/90',
        'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
        'disabled:pointer-events-none disabled:opacity-50',
        'transition-colors',
      )}
      {...props}
    />
  )
}

// AlertDialogCancel — cancel button (outline style)
type AlertDialogCancelProps = Omit<
  AlertDialogPrimitive.Close.Props,
  'className'
>

function AlertDialogCancel({ ...props }: AlertDialogCancelProps) {
  return (
    <AlertDialogPrimitive.Close
      data-slot="alert-dialog-cancel"
      className={cn(
        'inline-flex items-center justify-center rounded-lg border border-border bg-background px-4 py-2 text-sm font-medium',
        'hover:bg-muted hover:text-foreground',
        'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
        'disabled:pointer-events-none disabled:opacity-50',
        'transition-colors',
      )}
      {...props}
    />
  )
}

export {
  AlertDialog,
  AlertDialogTrigger,
  AlertDialogContent,
  AlertDialogHeader,
  AlertDialogFooter,
  AlertDialogTitle,
  AlertDialogDescription,
  AlertDialogAction,
  AlertDialogCancel,
}
