/**
 * Generates an array of numbers from start (inclusive) to end (exclusive)
 *
 * @param start - If end is provided, this is the starting number.
 *               If end is not provided, this is treated as the end number,
 *               and start is set to 0
 * @param end - The end number (exclusive). Optional parameter
 * @returns An array of sequential numbers
 *
 * @example
 * // Returns [0, 1, 2, 3, 4]
 * range(5)
 *
 * // Returns [2, 3, 4, 5]
 * range(2, 6)
 */
export function range(start: number, end?: number): number[] {
  if (end === undefined) {
    end = start
    start = 0
  }

  const result: number[] = []

  for (let i = start; i < end; i++) {
    result.push(i)
  }

  return result
}

/**
 * Returns a new array with elements sorted according to the comparison function
 * @param array The array to sort
 * @param compareFn Function used to determine the order of the elements
 * @returns A new sorted array
 */
export function toSorted<T>(array: T[], compareFn: (a: T, b: T) => number): T[] {
  return [...array].sort(compareFn)
}

/**
 * Splits an array into smaller chunks of a specified size
 *
 * @param array - The source array to be chunked
 * @param chunkSize - The size of each chunk
 * @returns A new array containing all chunks
 * @throws {Error} When chunkSize is less than or equal to 0
 *
 * @example
 * // Returns [[1, 2], [3, 4], [5]]
 * chunk([1, 2, 3, 4, 5], 2)
 */
export function chunk<T>(array: T[], chunkSize: number): T[][] {
  if (chunkSize <= 0) {
    throw new Error('Chunk size must be greater than zero')
  }

  const result: T[][] = []

  for (let i = 0; i < array.length; i += chunkSize) {
    const chunk = array.slice(i, i + chunkSize)
    result.push(chunk)
  }

  return result
}
