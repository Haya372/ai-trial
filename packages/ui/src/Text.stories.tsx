import type { Meta, StoryObj } from '@storybook/react'
import { Text } from './Text'

const meta = {
  title: 'Components/Text',
  component: Text,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
  argTypes: {
    variant: {
      control: 'select',
      options: ['h1', 'h2', 'h3', 'body', 'caption', 'code'],
    },
  },
} satisfies Meta<typeof Text>

export default meta
type Story = StoryObj<typeof meta>

export const Heading1: Story = {
  args: {
    variant: 'h1',
    children: 'Heading 1',
  },
}

export const Heading2: Story = {
  args: {
    variant: 'h2',
    children: 'Heading 2',
  },
}

export const Heading3: Story = {
  args: {
    variant: 'h3',
    children: 'Heading 3',
  },
}

export const Body: Story = {
  args: {
    variant: 'body',
    children: 'The quick brown fox jumps over the lazy dog.',
  },
}

export const Caption: Story = {
  args: {
    variant: 'caption',
    children: 'This is a caption or helper text.',
  },
}

export const Code: Story = {
  args: {
    variant: 'code',
    children: 'const greeting = "Hello, world!"',
  },
}
