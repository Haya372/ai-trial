import type { Meta, StoryObj } from '@storybook/react'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from './Card'
import { Button } from './Button'

const meta = {
  title: 'Components/Card',
  component: Card,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
} satisfies Meta<typeof Card>

export default meta
type Story = StoryObj<typeof meta>

export const Basic: Story = {
  render: () => (
    <Card>
      <CardContent>
        <p>This is a basic card with just content.</p>
      </CardContent>
    </Card>
  ),
}

export const WithHeaderAndFooter: Story = {
  render: () => (
    <Card>
      <CardHeader>
        <CardTitle>Card Title</CardTitle>
      </CardHeader>
      <CardContent>
        <p>Card body content goes here.</p>
      </CardContent>
      <CardFooter>
        <Button variant="primary">Action</Button>
      </CardFooter>
    </Card>
  ),
}

export const WithDescription: Story = {
  render: () => (
    <Card>
      <CardHeader>
        <CardTitle>Card with Description</CardTitle>
        <CardDescription>
          This is a supporting description that provides more context about the
          card content.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <p>Main content area for detailed information.</p>
      </CardContent>
      <CardFooter>
        <Button variant="outline">Cancel</Button>
        <Button variant="primary">Confirm</Button>
      </CardFooter>
    </Card>
  ),
}
