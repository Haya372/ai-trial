import type { Meta, StoryObj } from '@storybook/react'
import { Icon } from './Icon'

const meta = {
  title: 'Components/Icon',
  component: Icon,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
} satisfies Meta<typeof Icon>

export default meta
type Story = StoryObj<typeof meta>

const CircleIcon = () => (
  <svg
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    aria-hidden="true"
  >
    <circle cx="12" cy="12" r="10" />
  </svg>
)

export const ExtraSmall: Story = {
  render: () => (
    <Icon size="xs">
      <CircleIcon />
    </Icon>
  ),
}

export const Small: Story = {
  render: () => (
    <Icon size="sm">
      <CircleIcon />
    </Icon>
  ),
}

export const Medium: Story = {
  render: () => (
    <Icon size="md">
      <CircleIcon />
    </Icon>
  ),
}

export const Large: Story = {
  render: () => (
    <Icon size="lg">
      <CircleIcon />
    </Icon>
  ),
}

export const ExtraLarge: Story = {
  render: () => (
    <Icon size="xl">
      <CircleIcon />
    </Icon>
  ),
}

export const AllSizes: Story = {
  render: () => (
    <div className="flex items-center gap-4">
      <Icon size="xs">
        <CircleIcon />
      </Icon>
      <Icon size="sm">
        <CircleIcon />
      </Icon>
      <Icon size="md">
        <CircleIcon />
      </Icon>
      <Icon size="lg">
        <CircleIcon />
      </Icon>
      <Icon size="xl">
        <CircleIcon />
      </Icon>
    </div>
  ),
}
