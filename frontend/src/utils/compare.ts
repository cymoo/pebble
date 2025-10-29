/**
 * Performs a deep comparison between two values to determine if they are equivalent.
 * This function handles primitive types, functions, arrays, and objects.
 * Fork from: https://github.com/floating-ui/floating-ui/blob/master/packages/react-native/src/utils/deepEqual.ts
 *
 * @param a - The first value to compare.
 * @param b - The second value to compare.
 * @returns `true` if the values are deeply equal, otherwise `false`.
 *
 * @example
 * // Returns true
 * deepCompare({ a: 1, b: [2, 3] }, { a: 1, b: [2, 3] });
 *
 * // Returns false
 * deepCompare({ a: 1, b: [2, 3] }, { a: 1, b: [2, 4] });
 *
 * // Returns true
 * deepCompare(NaN, NaN);
 */
export function deepCompare(a: unknown, b: unknown): boolean {
  if (a === b) {
    return true
  }

  if (typeof a !== typeof b) {
    return false
  }

  if (typeof a === 'function' && typeof b === 'function' && a.toString() === b.toString()) {
    return true
  }

  if (Array.isArray(a) && Array.isArray(b)) {
    if (a.length !== b.length) {
      return false
    }

    const len = a.length
    for (let i = 0; i < len; i++) {
      if (!deepCompare(a[i], b[i])) {
        return false
      }
    }

    return true
  }

  if (a && b && typeof a === 'object' && typeof b === 'object') {
    const keysA = Object.keys(a)
    const keysB = Object.keys(b)

    if (keysA.length !== keysB.length) {
      return false
    }

    const len = keysA.length
    for (let i = 0; i < len; i++) {
      if (!{}.hasOwnProperty.call(b, keysA[i])) {
        return false
      }
    }

    for (let i = 0; i < len; i++) {
      const key = keysA[i]
      if (!deepCompare(a[key as keyof typeof a], b[key as keyof typeof b])) {
        return false
      }
    }

    return true
  }

  // NaN !== NaN
  return a !== a && b !== b
}
