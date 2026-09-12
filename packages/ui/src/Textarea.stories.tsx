import type { Meta, StoryObj } from '@storybook/react'
import { Textarea } from './Textarea'

const meta = {
  title: 'Components/Textarea',
  component: Textarea,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
  argTypes: {
    state: {
      control: 'select',
      options: ['default', 'error'],
    },
    disabled: { control: 'boolean' },
  },
} satisfies Meta<typeof Textarea>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  args: {
    rows: 4,
  },
}

export const Error: Story = {
  args: {
    rows: 4,
    state: 'error',
    'aria-invalid': true,
    defaultValue: 'Invalid content',
  },
}

export const Disabled: Story = {
  args: {
    rows: 4,
    disabled: true,
    defaultValue: 'Disabled textarea',
  },
}

export const WithPlaceholder: Story = {
  args: {
    rows: 4,
    placeholder: 'Enter your message here...',
  },
}

export const Resizable: Story = {
  args: {
    rows: 4,
    placeholder: 'This textarea can be resized (resize handle at bottom-right)',
  },
}
