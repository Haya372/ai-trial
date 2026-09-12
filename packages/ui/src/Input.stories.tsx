import type { Meta, StoryObj } from '@storybook/react'
import { Input } from './Input'

const meta = {
  title: 'Components/Input',
  component: Input,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
  argTypes: {
    size: {
      control: 'select',
      options: ['sm', 'md', 'lg'],
    },
    state: {
      control: 'select',
      options: ['default', 'error'],
    },
    disabled: { control: 'boolean' },
  },
} satisfies Meta<typeof Input>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  args: {
    placeholder: 'Enter text',
  },
}

export const Small: Story = {
  args: {
    placeholder: 'Small input',
    size: 'sm',
  },
}

export const Large: Story = {
  args: {
    placeholder: 'Large input',
    size: 'lg',
  },
}

export const Error: Story = {
  args: {
    placeholder: 'Error state',
    state: 'error',
    'aria-invalid': true,
  },
}

export const Disabled: Story = {
  args: {
    placeholder: 'Disabled input',
    disabled: true,
  },
}

export const WithValue: Story = {
  args: {
    defaultValue: 'Hello, world!',
  },
}
