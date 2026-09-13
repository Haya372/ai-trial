import type { Meta, StoryObj } from '@storybook/react'
import { List, ListItem } from './List'

const meta = {
  title: 'Components/List',
  component: List,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
} satisfies Meta<typeof List>

export default meta
type Story = StoryObj<typeof meta>

export const Unordered: Story = {
  render: () => (
    <List variant="unordered">
      <ListItem>First item</ListItem>
      <ListItem>Second item</ListItem>
      <ListItem>Third item</ListItem>
    </List>
  ),
}

export const Ordered: Story = {
  render: () => (
    <List variant="ordered">
      <ListItem>Step one</ListItem>
      <ListItem>Step two</ListItem>
      <ListItem>Step three</ListItem>
    </List>
  ),
}

export const None: Story = {
  render: () => (
    <List variant="none">
      <ListItem>No marker item</ListItem>
      <ListItem>Another item</ListItem>
      <ListItem>Yet another item</ListItem>
    </List>
  ),
}

export const WithIcons: Story = {
  render: () => (
    <List variant="none">
      <ListItem>
        <span className="inline-flex items-center gap-2">
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            aria-hidden="true"
          >
            <polyline points="20 6 9 17 4 12" />
          </svg>
          Completed task
        </span>
      </ListItem>
      <ListItem>
        <span className="inline-flex items-center gap-2">
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            aria-hidden="true"
          >
            <circle cx="12" cy="12" r="10" />
            <line x1="12" y1="8" x2="12" y2="12" />
            <line x1="12" y1="16" x2="12.01" y2="16" />
          </svg>
          Important notice
        </span>
      </ListItem>
      <ListItem>
        <span className="inline-flex items-center gap-2">
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            aria-hidden="true"
          >
            <path d="M5 12h14" />
            <path d="M12 5l7 7-7 7" />
          </svg>
          Navigate to next
        </span>
      </ListItem>
    </List>
  ),
}
