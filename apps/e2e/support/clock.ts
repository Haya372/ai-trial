export const TIMEZONE_ID = 'Asia/Tokyo'

// 月・週の途中かつ日中にして、前後の予定が月・週・日をまたがないようにする。
// 共有リンクの有効期限はサーバーが実時間で判定するため、実時間より未来の日付にしておく
export const FIXED_NOW = new Date('2030-06-12T10:00:00+09:00')

const JST_OFFSET_MS = 9 * 60 * 60 * 1000
const DAY_MS = 24 * 60 * 60 * 1000

type FixedDayTime = {
  /** FIXED_NOW の日付からの日数（前日なら -1） */
  dayOffset?: number
  hour: number
  minute?: number
}

/** FIXED_NOW の日付（JST）を基準に、指定した時刻を API が受け付ける UTC の RFC 3339 文字列で返す */
export function atFixedDay({ dayOffset = 0, hour, minute = 0 }: FixedDayTime): string {
  const jstMidnightMs = Math.floor((FIXED_NOW.getTime() + JST_OFFSET_MS) / DAY_MS) * DAY_MS - JST_OFFSET_MS
  const ms = jstMidnightMs + dayOffset * DAY_MS + (hour * 60 + minute) * 60 * 1000
  return new Date(ms).toISOString().replace('.000Z', 'Z')
}

const pad = (n: number) => String(n).padStart(2, '0')

function toJstParts(time: FixedDayTime) {
  const jst = new Date(new Date(atFixedDay(time)).getTime() + JST_OFFSET_MS)
  return {
    year: jst.getUTCFullYear(),
    month: jst.getUTCMonth() + 1,
    day: jst.getUTCDate(),
    hhmm: `${pad(jst.getUTCHours())}:${pad(jst.getUTCMinutes())}`,
  }
}

/** フォームの datetime-local 入力欄に入れる値（JST）を返す */
export function fixedDayInputValue(time: FixedDayTime): string {
  const { year, month, day, hhmm } = toJstParts(time)
  return `${year}-${pad(month)}-${pad(day)}T${hhmm}`
}

/** 予定詳細モーダルに表示される日時（JST）を返す */
export function fixedDayLabel(time: FixedDayTime): string {
  const { year, month, day, hhmm } = toJstParts(time)
  return `${year}年${month}月${day}日 ${hhmm}`
}
