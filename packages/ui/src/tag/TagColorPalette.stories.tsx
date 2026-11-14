import type { Meta, StoryObj } from '@storybook/react'
import { useState } from 'react'
import { TagBadge } from './TagBadge'
import { TagColorPalette } from './TagColorPalette'
import { TagColorSwatch } from './TagColorSwatch'
import { TAG_COLOR_KEYS, type TagColorKey } from './tagColorTokens'

const colorLabels: Record<TagColorKey, string> = {
  red: '赤',
  orange: 'オレンジ',
  yellow: '黄',
  green: '緑',
  teal: 'ティール',
  cyan: '水色',
  blue: '青',
  indigo: '藍',
  purple: '紫',
  pink: 'ピンク',
  brown: '茶',
  gray: 'グレー',
}

const meta = {
  title: 'Components/Tag/TagColorPalette',
  component: TagColorPalette,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
  args: {
    value: 'blue',
    onChange: () => {},
    colorLabels,
    'aria-label': '色',
  },
} satisfies Meta<typeof TagColorPalette>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  render: (args) => {
    const [value, setValue] = useState<TagColorKey>(args.value)
    return <TagColorPalette {...args} value={value} onChange={setValue} />
  },
}

export const Disabled: Story = {
  args: {
    disabled: true,
  },
}

// Shows every color as a badge and swatch so contrast and hue spacing can be reviewed at a glance.
export const AllColors: Story = {
  render: () => (
    <div className="flex flex-col gap-2">
      {TAG_COLOR_KEYS.map((color) => (
        <div key={color} className="flex items-center gap-2">
          <TagColorSwatch color={color} size="sm" />
          <TagColorSwatch color={color} />
          <TagBadge name={colorLabels[color]} color={color} size="sm" />
          <TagBadge name={`${colorLabels[color]}のタグ`} color={color} />
        </div>
      ))}
    </div>
  ),
}
