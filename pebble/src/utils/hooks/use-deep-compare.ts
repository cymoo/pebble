import { DependencyList, EffectCallback, useCallback, useEffect, useMemo, useRef } from 'react'

import { deepCompare } from '../compare.ts'

// from: https://github.com/sandiiarov/use-deep-compare
export function useDeepCompareMemoize(dependencies: DependencyList): DependencyList {
  const dependenciesRef = useRef<DependencyList>(dependencies)
  const signalRef = useRef<number>(0)

  if (!deepCompare(dependencies, dependenciesRef.current)) {
    dependenciesRef.current = dependencies
    signalRef.current += 1
  }

  return useMemo(() => dependenciesRef.current, [signalRef.current])
}

export function useDeepCompareMemo<T>(factory: () => T, dependencies: DependencyList): T {
  return useMemo(factory, useDeepCompareMemoize(dependencies))
}

export function useDeepCompareEffect(effect: EffectCallback, dependencies: DependencyList) {
  useEffect(effect, useDeepCompareMemoize(dependencies))
}

// eslint-disable-next-line @typescript-eslint/no-unsafe-function-type
export function useDeepCompareCallback<T extends Function>(
  callback: T,
  dependencies: DependencyList,
) {
  return useCallback(callback, useDeepCompareMemoize(dependencies))
}
