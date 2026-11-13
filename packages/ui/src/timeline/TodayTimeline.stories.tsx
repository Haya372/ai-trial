import type { Meta, StoryObj } from '@storybook/react'
import { TodayTimeline } from './TodayTimeline'

const at = (hours: number, minutes = 0) =>
  new Date(2026, 10, 14, hours, minutes)

const labels = {
  title: '今日の予定',
  noEventsToday: '今日の予定はありません',
  noMoreEventsToday: '今日の予定はもうありません',
  fromPreviousDay: '前日から',
  untilNextDay: '翌日まで',
  minutesUntil: (minutes: number) => `あと${minutes}分`,
  hoursMinutesUntil: (hours: number, minutes: number) =>
    `あと${hours}時間${minutes}分`,
  freeTime: (minutes: number) => `${minutes}分の空き`,
}

const sampleEvents = [
  { id: '1', title: '朝会', start: at(9), end: at(9, 30) },
  { id: '2', title: 'デザインレビュー', start: at(10), end: at(11) },
  {
    id: '3',
    title:
      '四半期ロードマップのすり合わせと来期の採用計画に関するディスカッション',
    start: at(13),
    end: at(14, 30),
  },
  { id: '4', title: '1on1', start: at(16), end: at(16, 30) },
]

const meta = {
  title: 'Calendar/TodayTimeline',
  component: TodayTimeline,
  parameters: {
    layout: 'padded',
  },
  decorators: [
    (Story) => (
      <div style={{ width: 280 }}>
        <Story />
      </div>
    ),
  ],
  tags: ['autodocs'],
  args: {
    events: sampleEvents,
    now: at(9, 40),
    labels,
    onEventClick: (event) => console.log('Event clicked:', event),
  },
} satisfies Meta<typeof TodayTimeline>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {}

export const NextEventOverAnHour: Story = {
  args: {
    now: at(11, 15),
  },
}

export const NoEvents: Story = {
  args: {
    events: [],
  },
}

export const NoMoreEvents: Story = {
  args: {
    now: at(18),
  },
}

export const OverlappingEvents: Story = {
  args: {
    events: [
      { id: '1', title: 'ミーティングA', start: at(10), end: at(11, 30) },
      { id: '2', title: 'ミーティングB', start: at(10, 30), end: at(12) },
      { id: '3', title: 'ランチ', start: at(12, 30), end: at(13, 30) },
    ],
    now: at(10, 45),
  },
}

export const SpanningDays: Story = {
  args: {
    events: [
      {
        id: '1',
        title: '夜間メンテナンス',
        start: new Date(2026, 10, 13, 22),
        end: at(1),
      },
      { id: '2', title: '定例', start: at(10), end: at(11) },
      {
        id: '3',
        title: 'リリース作業',
        start: at(23),
        end: new Date(2026, 10, 15, 2),
      },
    ],
    now: at(0, 30),
  },
}
