import type { Meta, StoryObj } from '@storybook/react'
import { MonthCalendar } from './MonthCalendar'

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
    start: new Date(2026, 8, 20, 10, 0),
    end: new Date(2026, 8, 20, 10, 30),
  },
]

const meta = {
  title: 'Calendar/MonthCalendar',
  component: MonthCalendar,
  parameters: {
    layout: 'fullscreen',
  },
  tags: ['autodocs'],
  args: {
    events: sampleEvents,
    currentDate: new Date(2026, 8, 1),
    onEventClick: (event) => console.log('Event clicked:', event),
    onDateChange: (date) => console.log('Date changed:', date),
  },
} satisfies Meta<typeof MonthCalendar>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {}

export const NoEvents: Story = {
  args: {
    events: [],
  },
}

export const ManyEvents: Story = {
  args: {
    events: Array.from({ length: 20 }, (_, i) => ({
      id: String(i),
      title: `イベント${i + 1}`,
      start: new Date(2026, 8, (i % 28) + 1, 9, 0),
      end: new Date(2026, 8, (i % 28) + 1, 10, 0),
    })),
  },
}
