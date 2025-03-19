import { useCallback, useRef, useState } from 'react'

import { useUnmount } from './use-unmount.ts'

// https://github.com/alibaba/hooks/blob/master/packages/hooks/src/useRafState/index.ts
export function useRafState<S>(initialState: S | (() => S)) {
  const ref = useRef(0)
  const [state, setState] = useState(initialState)

  const setRafState = useCallback((value: S | ((prevState: S) => S)) => {
    cancelAnimationFrame(ref.current)

    ref.current = requestAnimationFrame(() => {
      setState(value)
    })
  }, [])

  useUnmount(() => {
    cancelAnimationFrame(ref.current)
  })

  return [state, setRafState] as const
}
