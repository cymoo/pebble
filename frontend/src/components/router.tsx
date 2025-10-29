import { ReactNode, createContext, useContext, useEffect, useState } from 'react'
import { Location, NavigateFunction, useLocation, useNavigate } from 'react-router'

const BackgroundLocationContext = createContext<string | null>(null)

export function BackgroundLocationProvider({ children }: { children: ReactNode }) {
  let location = useLocation()

  const state = location.state as { backgroundLocation?: Location } | null

  if (state?.backgroundLocation) {
    location = state.backgroundLocation
  }

  const value = JSON.stringify(location)

  return <BackgroundLocationContext value={value}>{children}</BackgroundLocationContext>
}

export function useBackgroundLocation(): Location {
  const bgLocation = useContext(BackgroundLocationContext)
  if (bgLocation === null) {
    throw Error('`useBackgroundLocation` cannot be called outside of its provider')
  }
  return JSON.parse(bgLocation) as Location
}

const StableNavigateContext = createContext<NavigateFunction | null>(null)

export function StableNavigateProvider({ children }: { children: ReactNode }) {
  const originalNavigate = useNavigate()
  // use a function as useState's initial to prevent `history` warning:
  // navigate cannot be called during rendering.
  const [navigate] = useState(() => originalNavigate)

  useEffect(() => {
    window.navigate = navigate
  }, [navigate])

  return <StableNavigateContext value={navigate}>{children}</StableNavigateContext>
}

// https://github.com/remix-run/react-router/issues/7634
export function useStableNavigate(): NavigateFunction {
  const func = useContext(StableNavigateContext)
  if (func === null) {
    throw Error('`useStableNavigate` cannot be called outside of its provider')
  }
  return func
}

declare global {
  interface Window {
    navigate: NavigateFunction
  }
}
