import { useCallback, useLayoutEffect, useRef } from 'react'

// a user-land implementation of `useEvent`
// https://github.com/reactjs/rfcs/pull/220
// https://github.com/reactjs/rfcs/blob/useevent/text/0000-useevent.md

// https://stackoverflow.com/questions/67713566/how-does-the-never-type-work-in-typescript
export function useEvent<T extends (...args: Parameters<T>) => ReturnType<T>>(handler: T): T {
  const handlerRef = useRef<T>(null)

  // In a real implementation, this would run before layout effects
  useLayoutEffect(() => {
    handlerRef.current = handler
  })

  return useCallback((...args: Parameters<T>) => {
    // In a real implementation, this would throw if called during render
    const fn = handlerRef.current
    if (!fn) {
      throw new Error('useEvent cannot be called during render')
    }
    return fn(...args)
  }, []) as T
}
