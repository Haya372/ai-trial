import type { SelectRootProps } from '@base-ui/react/select'
import type { Meta, StoryObj } from '@storybook/react'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectItemText,
  SelectTrigger,
  SelectValue,
} from './Select'

// Concrete wrapper to give Storybook a non-generic component to infer from
function SelectDemo(props: SelectRootProps<string>) {
  return <Select {...props} />
}

const meta = {
  title: 'Components/Select',
  component: SelectDemo,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
} satisfies Meta<typeof SelectDemo>

export default meta
type Story = StoryObj<typeof meta>

const FruitOptions = () => (
  <SelectContent>
    <SelectItem value="apple">
      <SelectItemText>Apple</SelectItemText>
    </SelectItem>
    <SelectItem value="banana">
      <SelectItemText>Banana</SelectItemText>
    </SelectItem>
    <SelectItem value="cherry">
      <SelectItemText>Cherry</SelectItemText>
    </SelectItem>
    <SelectItem value="grape">
      <SelectItemText>Grape</SelectItemText>
    </SelectItem>
    <SelectItem value="mango">
      <SelectItemText>Mango</SelectItemText>
    </SelectItem>
  </SelectContent>
)

export const Default: Story = {
  render: () => (
    <SelectDemo>
      <SelectTrigger>
        <SelectValue placeholder="Select a fruit" />
      </SelectTrigger>
      <FruitOptions />
    </SelectDemo>
  ),
}

export const Small: Story = {
  render: () => (
    <SelectDemo>
      <SelectTrigger size="sm">
        <SelectValue placeholder="Select a fruit" />
      </SelectTrigger>
      <FruitOptions />
    </SelectDemo>
  ),
}

export const Large: Story = {
  render: () => (
    <SelectDemo>
      <SelectTrigger size="lg">
        <SelectValue placeholder="Select a fruit" />
      </SelectTrigger>
      <FruitOptions />
    </SelectDemo>
  ),
}

export const Disabled: Story = {
  render: () => (
    <SelectDemo disabled>
      <SelectTrigger>
        <SelectValue placeholder="Select a fruit" />
      </SelectTrigger>
      <FruitOptions />
    </SelectDemo>
  ),
}

export const WithDefaultValue: Story = {
  render: () => (
    <SelectDemo defaultValue="apple">
      <SelectTrigger>
        <SelectValue placeholder="Select a fruit" />
      </SelectTrigger>
      <FruitOptions />
    </SelectDemo>
  ),
}
