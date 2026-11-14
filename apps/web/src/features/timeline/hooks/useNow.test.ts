import { act, renderHook } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { useNow } from './useNow'

function setVisibility(state: DocumentVisibilityState) {
  Object.defineProperty(document, 'visibilityState', {
    configurable: true,
    get: () => state,
  })
  document.dispatchEvent(new Event('visibilitychange'))
}

describe('useNow', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date(2026, 10, 14, 10, 0, 30))
  })
  afterEach(() => {
    vi.useRealTimers()
  })

  it('初回は現在時刻を返す', () => {
    const { result } = renderHook(() => useNow())
    expect(result.current).toEqual(new Date(2026, 10, 14, 10, 0, 30))
  })

  it('分の途中から開始しても、次の分の境界で更新する', () => {
    const { result } = renderHook(() => useNow())

    act(() => {
      vi.advanceTimersByTime(29_999)
    })
    expect(result.current).toEqual(new Date(2026, 10, 14, 10, 0, 30))

    act(() => {
      vi.advanceTimersByTime(1)
    })
    expect(result.current).toEqual(new Date(2026, 10, 14, 10, 1, 0))
  })

  it('更新後も次の分の境界ごとに更新し続ける', () => {
    const { result } = renderHook(() => useNow())

    act(() => {
      vi.advanceTimersByTime(30_000 + 60_000)
    })
    expect(result.current).toEqual(new Date(2026, 10, 14, 10, 2, 0))
  })

  it('タブが再表示されたとき現在時刻で更新する', () => {
    const { result } = renderHook(() => useNow())

    // 非表示中にタイマーが発火しないまま時刻だけ進んだ状況
    vi.setSystemTime(new Date(2026, 10, 15, 0, 5, 10))
    act(() => {
      setVisibility('visible')
    })

    expect(result.current).toEqual(new Date(2026, 10, 15, 0, 5, 10))
  })

  it('タブが非表示になったときは更新しない', () => {
    const { result } = renderHook(() => useNow())

    vi.setSystemTime(new Date(2026, 10, 14, 10, 0, 40))
    act(() => {
      setVisibility('hidden')
    })

    expect(result.current).toEqual(new Date(2026, 10, 14, 10, 0, 30))
  })

  it('アンマウント後はタイマーを残さない', () => {
    const { unmount } = renderHook(() => useNow())
    unmount()
    expect(vi.getTimerCount()).toBe(0)
  })
})
