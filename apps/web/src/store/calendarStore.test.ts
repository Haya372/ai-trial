import { beforeEach, describe, expect, it } from 'vitest'
import { useCalendarStore } from './calendarStore'

describe('calendarStore', () => {
  beforeEach(() => {
    // Zustandのストアをリセットする
    useCalendarStore.setState({
      view: 'month',
      currentDate: new Date(2026, 8, 1), // 2026-09-01
    })
  })

  describe('初期状態', () => {
    it('view の初期値は "month" である', () => {
      const { view } = useCalendarStore.getState()
      expect(view).toBe('month')
    })

    it('currentDate の初期値は本日の日付である', () => {
      // ストアをデフォルト状態にリセット
      useCalendarStore.setState({
        view: 'month',
        currentDate: new Date(),
      })
      const { currentDate } = useCalendarStore.getState()
      const today = new Date()
      expect(currentDate.getFullYear()).toBe(today.getFullYear())
      expect(currentDate.getMonth()).toBe(today.getMonth())
      expect(currentDate.getDate()).toBe(today.getDate())
    })
  })

  describe('setView', () => {
    it('"week" を渡すと view が "week" になる', () => {
      const { setView } = useCalendarStore.getState()
      setView('week')
      expect(useCalendarStore.getState().view).toBe('week')
    })

    it('"month" を渡すと view が "month" になる', () => {
      useCalendarStore.setState({ view: 'week' })
      const { setView } = useCalendarStore.getState()
      setView('month')
      expect(useCalendarStore.getState().view).toBe('month')
    })
  })

  describe('setCurrentDate', () => {
    it('新しい日付を渡すと currentDate が更新される', () => {
      const { setCurrentDate } = useCalendarStore.getState()
      const newDate = new Date(2026, 10, 15) // 2026-11-15
      setCurrentDate(newDate)
      expect(useCalendarStore.getState().currentDate).toEqual(newDate)
    })
  })
})
