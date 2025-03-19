import { useLayoutEffect, useState } from 'react'

export function useWindowSize() {
  const [size, setSize] = useState<{
    width: number
    height: number
  }>({
    width: window.innerWidth,
    height: window.innerHeight,
  })
  // avoid flicking by using `useLayoutEffect` instead of `useEffect`
  useLayoutEffect(() => {
    function handleResize() {
      setSize({
        width: window.innerWidth,
        height: window.innerHeight,
      })
    }
    window.addEventListener('resize', handleResize)
    return () => {
      window.removeEventListener('resize', handleResize)
    }
  }, [])
  return size
}
