import type { Meta, StoryObj } from '@storybook/react'
import { WeekCalendar } from './WeekCalendar'

const sampleEvents = [
  {
    id: '1',
    title: 'デザインレビュー',
    start: new Date(2026, 8, 13, 10, 0),
    end: new Date(2026, 8, 13, 11, 0),
  },
  {
    id: '2',
    title: 'スプリント計画',
    start: new Date(2026, 8, 15, 14, 0),
    end: new Date(2026, 8, 15, 16, 0),
  },
  {
    id: '3',
    title: '1on1',
    start: new Date(2026, 8, 16, 10, 0),
    end: new Date(2026, 8, 16, 10, 30),
  },
]

const meta = {
  title: 'Calendar/WeekCalendar',
  component: WeekCalendar,
  parameters: {
    layout: 'fullscreen',
  },
  tags: ['autodocs'],
  args: {
    events: sampleEvents,
    currentDate: new Date(2026, 8, 13),
    onEventClick: (event) => console.log('Event clicked:', event),
    onDateChange: (date) => console.log('Date changed:', date),
  },
} satisfies Meta<typeof WeekCalendar>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {}

export const NoEvents: Story = {
  args: {
    events: [],
  },
}

export const OverlappingEvents: Story = {
  args: {
    events: [
      {
        id: '1',
        title: 'ミーティングA',
        start: new Date(2026, 8, 14, 10, 0),
        end: new Date(2026, 8, 14, 11, 30),
      },
      {
        id: '2',
        title: 'ミーティングB',
        start: new Date(2026, 8, 14, 10, 30),
        end: new Date(2026, 8, 14, 12, 0),
      },
    ],
  },
}
