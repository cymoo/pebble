/**
 * Filter out null values and concatenate the strings with spaces.
 * @example
 * cx('foo', undefined, 'bar') === 'foo bar'
 * cx('foo', {bar: true}) === 'foo bar'
 * cx('foo', {bar: false}) === 'foo'
 */
export function cx(...args: (string | Record<string, boolean> | null | undefined)[]): string {
  return args
    .map((arg) => {
      if (typeof arg === 'object') {
        if (arg === null) {
          return null
        } else {
          return Object.entries(arg)
            .filter(([, value]) => value)
            .map(([key]) => key)
            .join(' ')
        }
      } else {
        return arg
      }
    })
    .filter((arg) => !!arg)
    .join(' ')
}
