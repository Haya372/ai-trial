import type { Meta, StoryObj } from '@storybook/react'
import { Spinner } from './Spinner'

const meta = {
  title: 'Components/Spinner',
  component: Spinner,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
  argTypes: {
    size: {
      control: 'select',
      options: ['sm', 'md', 'lg'],
    },
    label: { control: 'text' },
  },
} satisfies Meta<typeof Spinner>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  args: {
    label: 'Loading',
  },
}

export const Small: Story = {
  args: {
    size: 'sm',
    label: 'Loading',
  },
}

export const Large: Story = {
  args: {
    size: 'lg',
    label: 'Loading',
  },
}

export const Decorative: Story = {
  args: {},
}
