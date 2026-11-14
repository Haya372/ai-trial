import { useEffect, useState } from 'react'

const MINUTE_MS = 60_000

// ローカル時刻の分の境界ごとに現在時刻を更新する。
// setInterval だとタイマーの遅延が累積して境界からずれるため、毎回次の境界へ setTimeout し直す
export function useNow(): Date {
  const [now, setNow] = useState(() => new Date())

  useEffect(() => {
    let timer: ReturnType<typeof setTimeout>

    function schedule() {
      timer = setTimeout(
        () => {
          setNow(new Date())
          schedule()
        },
        MINUTE_MS - (Date.now() % MINUTE_MS),
      )
    }

    // 非表示中のタブはブラウザがタイマーを間引くため、再表示時に追いつかせる
    function handleVisibilityChange() {
      if (document.visibilityState !== 'visible') return
      clearTimeout(timer)
      setNow(new Date())
      schedule()
    }

    schedule()
    document.addEventListener('visibilitychange', handleVisibilityChange)
    return () => {
      clearTimeout(timer)
      document.removeEventListener('visibilitychange', handleVisibilityChange)
    }
  }, [])

  return now
}
