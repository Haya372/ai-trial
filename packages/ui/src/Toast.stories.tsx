import type { Meta, StoryObj } from '@storybook/react'
import { Button } from './Button'
import { toast, Toaster } from './Toast'

function ToastDemo({
  variant,
}: {
  variant?: 'default' | 'success' | 'warning' | 'destructive'
}) {
  const messages = {
    default: { title: 'Notification', description: 'This is a default toast.' },
    success: { title: 'Saved!', description: 'Your changes have been saved.' },
    warning: {
      title: 'Warning',
      description: 'Please review before continuing.',
    },
    destructive: { title: 'Error', description: 'Something went wrong.' },
  }

  const resolved = variant ?? 'default'
  const { title, description } = messages[resolved]

  return (
    <>
      <Toaster />
      <Button
        variant="outline"
        onClick={() => toast({ title, description, variant: resolved })}
      >
        Show {resolved} toast
      </Button>
    </>
  )
}

const meta = {
  title: 'Components/Toast',
  component: ToastDemo,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
} satisfies Meta<typeof ToastDemo>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = { args: { variant: 'default' } }
export const Success: Story = { args: { variant: 'success' } }
export const Warning: Story = { args: { variant: 'warning' } }
export const Destructive: Story = { args: { variant: 'destructive' } }

export const AllVariants: Story = {
  render: () => (
    <>
      <Toaster />
      <div className="flex flex-wrap gap-2">
        {(['default', 'success', 'warning', 'destructive'] as const).map(
          (variant) => (
            <Button
              key={variant}
              variant="outline"
              onClick={() =>
                toast({
                  title: variant.charAt(0).toUpperCase() + variant.slice(1),
                  description: `This is a ${variant} toast.`,
                  variant,
                })
              }
            >
              {variant}
            </Button>
          ),
        )}
      </div>
    </>
  ),
}
