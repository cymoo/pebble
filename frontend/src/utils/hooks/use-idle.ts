import { useEffect, useRef } from 'react'

import { useLatest } from './use-latest.ts'

// NOTE: Frequent calls to `mousemove` may impact performance.
// https://stackoverflow.com/a/31751827
const events = [
  // 'mousemove',
  'mousedown',
  'touchstart',
  // 'touchmove',
  'click',
  'keydown',
  'scroll',
] as const

export function useIdle(timeout: number | null, callback: () => void) {
  const fnRef = useLatest(callback)
  const tidRef = useRef<number | undefined>(undefined)

  useEffect(() => {
    if (!timeout) return

    const reset = () => {
      clearTimeout(tidRef.current)
      tidRef.current = window.setTimeout(fnRef.current, timeout * 1000)
    }

    for (const event of events) {
      window.addEventListener(event, reset, true)
    }
    return () => {
      clearTimeout(tidRef.current)
      for (const event of events) {
        window.removeEventListener(event, reset, true)
      }
    }
  }, [fnRef, timeout])
}
