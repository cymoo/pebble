import { useRef } from 'react'

// https://github.com/alibaba/hooks/tree/master/packages/hooks/src/usePrevious
export type ShouldUpdateFunc<T> = (prev: T | undefined, next: T) => boolean

const defaultShouldUpdate = <T>(a?: T, b?: T) => !Object.is(a, b)

export function usePrevious<T>(
  state: T,
  shouldUpdate: ShouldUpdateFunc<T> = defaultShouldUpdate,
): T | undefined {
  const prevRef = useRef<T>(undefined)
  const curRef = useRef<T>(undefined)

  if (shouldUpdate(curRef.current, state)) {
    prevRef.current = curRef.current
    curRef.current = state
  }

  return prevRef.current
}
