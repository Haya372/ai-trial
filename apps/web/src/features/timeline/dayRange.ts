// 1日の区切りはユーザーのタイムゾーンに依存するため、クライアントで計算する（ADR-036）。
// 24時間の加算ではなく日付の加算にして、夏時間の切り替え日も翌日0時に揃える
export function getDayRange(date: Date): { startAt: Date; endAt: Date } {
  return {
    startAt: new Date(date.getFullYear(), date.getMonth(), date.getDate()),
    endAt: new Date(date.getFullYear(), date.getMonth(), date.getDate() + 1),
  }
}
