import type { Meta, StoryObj } from '@storybook/react'
import { Loading } from './Loading'

const meta = {
  title: 'Components/Loading',
  component: Loading,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
  argTypes: {
    label: { control: 'text' },
  },
} satisfies Meta<typeof Loading>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  args: {
    label: 'Loading...',
  },
}

export const WithoutLabel: Story = {
  args: {},
}
