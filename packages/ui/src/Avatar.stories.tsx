import type { Meta, StoryObj } from '@storybook/react'
import { Avatar, AvatarFallback, AvatarImage } from './Avatar'

const meta = {
  title: 'Components/Avatar',
  component: Avatar,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
} satisfies Meta<typeof Avatar>

export default meta
type Story = StoryObj<typeof meta>

export const WithImage: Story = {
  render: () => (
    <Avatar>
      <AvatarImage src="https://github.com/shadcn.png" alt="User avatar" />
      <AvatarFallback>CN</AvatarFallback>
    </Avatar>
  ),
}

export const WithFallback: Story = {
  render: () => (
    <Avatar>
      <AvatarImage src="" alt="User avatar" />
      <AvatarFallback>AB</AvatarFallback>
    </Avatar>
  ),
}

export const Small: Story = {
  render: () => (
    <Avatar size="sm">
      <AvatarImage src="" alt="User avatar" />
      <AvatarFallback>SM</AvatarFallback>
    </Avatar>
  ),
}

export const Large: Story = {
  render: () => (
    <Avatar size="lg">
      <AvatarImage src="" alt="User avatar" />
      <AvatarFallback>LG</AvatarFallback>
    </Avatar>
  ),
}
