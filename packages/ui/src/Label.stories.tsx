import type { Meta, StoryObj } from '@storybook/react'
import { Label } from './Label'

const meta = {
  title: 'Components/Label',
  component: Label,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
} satisfies Meta<typeof Label>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  args: {
    children: 'Email address',
    htmlFor: 'email',
  },
}

export const WithInput: Story = {
  render: () => (
    <div className="flex flex-col gap-1.5">
      <Label htmlFor="username">Username</Label>
      <input
        id="username"
        type="text"
        placeholder="Enter your username"
        className="h-9 rounded-lg border border-input px-3 text-sm outline-none focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/50"
      />
    </div>
  ),
}

export const Required: Story = {
  render: () => (
    <div className="flex flex-col gap-1.5">
      <Label htmlFor="required-field">
        Full name <span className="text-destructive">*</span>
      </Label>
      <input
        id="required-field"
        type="text"
        placeholder="Enter your full name"
        className="h-9 rounded-lg border border-input px-3 text-sm outline-none focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/50"
      />
    </div>
  ),
}
