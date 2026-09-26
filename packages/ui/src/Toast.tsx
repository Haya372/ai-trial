import { cva, type VariantProps } from 'class-variance-authority'
import { toast as sonnerToast, Toaster as SonnerToaster } from 'sonner'

const toastVariants = cva('', {
  variants: {
    variant: {
      default: '',
      success: '',
      warning: '',
      destructive: '',
    },
  },
  defaultVariants: {
    variant: 'default',
  },
})

type ToastVariant = NonNullable<VariantProps<typeof toastVariants>['variant']>

const toastTypeMap: Record<ToastVariant, keyof typeof sonnerToast> = {
  default: 'message',
  success: 'success',
  warning: 'warning',
  destructive: 'error',
}

type ToastOptions = {
  title: string
  description?: string
  duration?: number
}

function toast(options: ToastOptions & { variant?: ToastVariant }) {
  const { title, description, duration, variant = 'default' } = options
  const method = toastTypeMap[variant]
  ;(sonnerToast[method] as (title: string, opts?: object) => void)(title, {
    description,
    duration,
  })
}

toast.success = (title: string, opts?: Omit<ToastOptions, 'title'>) =>
  toast({ ...opts, title, variant: 'success' })
toast.warning = (title: string, opts?: Omit<ToastOptions, 'title'>) =>
  toast({ ...opts, title, variant: 'warning' })
toast.error = (title: string, opts?: Omit<ToastOptions, 'title'>) =>
  toast({ ...opts, title, variant: 'destructive' })

type ToasterProps = {
  position?: React.ComponentProps<typeof SonnerToaster>['position']
  richColors?: boolean
  closeButton?: boolean
  duration?: number
}

function Toaster({
  position = 'top-center',
  richColors = true,
  closeButton = true,
  duration,
}: ToasterProps) {
  return (
    <SonnerToaster
      position={position}
      richColors={richColors}
      closeButton={closeButton}
      duration={duration}
    />
  )
}

export { toast, Toaster }
