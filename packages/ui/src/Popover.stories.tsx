import type { Meta, StoryObj } from '@storybook/react'
import { Button } from './Button'
import { Popover, PopoverContent, PopoverTrigger } from './Popover'

const meta = {
  title: 'Components/Popover',
  component: Popover,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
} satisfies Meta<typeof Popover>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  render: () => (
    <Popover>
      <PopoverTrigger render={<Button variant="outline">Open</Button>} />
      <PopoverContent>This is a popover</PopoverContent>
    </Popover>
  ),
}

export const Top: Story = {
  render: () => (
    <Popover>
      <PopoverTrigger render={<Button variant="outline">Top</Button>} />
      <PopoverContent side="top">Popover on top</PopoverContent>
    </Popover>
  ),
}
