import type { Meta, StoryObj } from '@storybook/react'
import { Tooltip, TooltipTrigger, TooltipContent } from './Tooltip'
import { Button } from './Button'

const meta = {
  title: 'Components/Tooltip',
  component: Tooltip,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
} satisfies Meta<typeof Tooltip>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  render: () => (
    <Tooltip>
      <TooltipTrigger render={<Button variant="outline">Hover me</Button>} />
      <TooltipContent>This is a tooltip</TooltipContent>
    </Tooltip>
  ),
}

export const Top: Story = {
  render: () => (
    <Tooltip>
      <TooltipTrigger render={<Button variant="outline">Top</Button>} />
      <TooltipContent side="top">Tooltip on top</TooltipContent>
    </Tooltip>
  ),
}

export const Bottom: Story = {
  render: () => (
    <Tooltip>
      <TooltipTrigger render={<Button variant="outline">Bottom</Button>} />
      <TooltipContent side="bottom">Tooltip on bottom</TooltipContent>
    </Tooltip>
  ),
}

export const Left: Story = {
  render: () => (
    <Tooltip>
      <TooltipTrigger render={<Button variant="outline">Left</Button>} />
      <TooltipContent side="left">Tooltip on left</TooltipContent>
    </Tooltip>
  ),
}

export const Right: Story = {
  render: () => (
    <Tooltip>
      <TooltipTrigger render={<Button variant="outline">Right</Button>} />
      <TooltipContent side="right">Tooltip on right</TooltipContent>
    </Tooltip>
  ),
}
