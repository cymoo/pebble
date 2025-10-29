import { useCallback, useEffect, useRef } from 'react'

import { useLatest } from './use-latest.ts'

export function useIsUnmounted(): () => boolean {
  const ref = useRef(false)

  useEffect(() => {
    ref.current = false
    return () => {
      ref.current = true
    }
  }, [])

  return useCallback(() => ref.current, [])
}

export function useUnmount(fn: () => void) {
  const fnRef = useLatest(fn)

  useEffect(
    () => () => {
      fnRef.current()
    },
    [fnRef],
  )
}
