import type { Meta, StoryObj } from '@storybook/react'
import { Divider } from './Divider'

const meta = {
  title: 'Components/Divider',
  component: Divider,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
  argTypes: {
    orientation: {
      control: 'select',
      options: ['horizontal', 'vertical'],
    },
  },
} satisfies Meta<typeof Divider>

export default meta
type Story = StoryObj<typeof meta>

export const Horizontal: Story = {
  render: (args) => (
    <div className="w-80">
      <p className="text-sm">Above the divider</p>
      <Divider {...args} />
      <p className="text-sm">Below the divider</p>
    </div>
  ),
  args: {
    orientation: 'horizontal',
  },
}

export const Vertical: Story = {
  render: (args) => (
    <div className="flex h-10 items-center gap-2">
      <span className="text-sm">Left</span>
      <Divider {...args} />
      <span className="text-sm">Right</span>
    </div>
  ),
  args: {
    orientation: 'vertical',
  },
}

export const InContent: Story = {
  render: () => (
    <div className="w-80 rounded-lg border border-border p-4">
      <p className="text-sm font-medium">Section A</p>
      <p className="text-sm text-muted-foreground">Content for section A.</p>
      <Divider />
      <p className="text-sm font-medium">Section B</p>
      <p className="text-sm text-muted-foreground">Content for section B.</p>
    </div>
  ),
}
