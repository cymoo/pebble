import { Ref, RefCallback, useMemo } from 'react'

// from: https://github.com/floating-ui/floating-ui/blob/master/packages/react/src/hooks/useMergeRefs.ts
export function useMergeRefs<T>(refs: (Ref<T> | undefined)[]): RefCallback<T> | null {
  return useMemo(() => {
    if (refs.every((ref) => ref == null)) {
      return null
    }

    return (value) => {
      refs.forEach((ref) => {
        if (typeof ref === 'function') {
          ref(value)
        } else if (ref != null) {
          ref.current = value
        }
      })
    }
  }, refs)
}
