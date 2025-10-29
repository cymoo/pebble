import { useEffect, useState } from 'react'

// https://caniuse.com/?search=window.visualViewport
function tryGetViewports(): { width: number; height: number } {
  if (window.visualViewport) {
    return {
      width: window.visualViewport.width,
      height: window.visualViewport.height,
    }
  } else {
    // when not available, use viewport as a fallback
    return {
      width: Math.max(document.documentElement.clientWidth || 0, window.innerWidth || 0),
      height: Math.max(document.documentElement.clientHeight || 0, window.innerHeight || 0),
    }
  }
}

export function useVisualViewportWhenPossible(): { width: number; height: number } {
  const [state, setState] = useState(tryGetViewports)
  useEffect(() => {
    const handleResize = () => {
      setState(tryGetViewports())
    }
    if (window.visualViewport) {
      window.visualViewport.addEventListener('resize', handleResize)
    }
    return () => {
      if (window.visualViewport) {
        window.visualViewport.removeEventListener('resize', handleResize)
      }
    }
  }, [])
  return state
}
