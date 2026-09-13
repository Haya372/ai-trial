import { Avatar as AvatarPrimitive } from '@base-ui/react/avatar'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from 'cn'

const avatarVariants = cva(
  'relative flex shrink-0 overflow-hidden rounded-full',
  {
    variants: {
      size: {
        sm: 'h-8 w-8',
        md: 'h-10 w-10',
        lg: 'h-14 w-14',
      },
    },
    defaultVariants: {
      size: 'md',
    },
  },
)

type AvatarProps = Omit<AvatarPrimitive.Root.Props, 'className'> &
  VariantProps<typeof avatarVariants>

function Avatar({ size, ...props }: AvatarProps) {
  return (
    <AvatarPrimitive.Root
      data-slot="avatar"
      className={cn(avatarVariants({ size }))}
      {...props}
    />
  )
}

type AvatarImageProps = Omit<AvatarPrimitive.Image.Props, 'className'>

function AvatarImage({ ...props }: AvatarImageProps) {
  return (
    <AvatarPrimitive.Image
      data-slot="avatar-image"
      className={cn('aspect-square h-full w-full object-cover')}
      {...props}
    />
  )
}

type AvatarFallbackProps = Omit<AvatarPrimitive.Fallback.Props, 'className'>

function AvatarFallback({ ...props }: AvatarFallbackProps) {
  return (
    <AvatarPrimitive.Fallback
      data-slot="avatar-fallback"
      className={cn(
        'flex h-full w-full items-center justify-center rounded-full bg-muted text-sm font-medium text-muted-foreground',
      )}
      {...props}
    />
  )
}

export { Avatar, AvatarImage, AvatarFallback }
