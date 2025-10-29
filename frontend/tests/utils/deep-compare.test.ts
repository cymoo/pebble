import { deepCompare } from '../../src/utils/compare'

test('deep compare', () => {
  let inputs = [
    [null, null, true],
    [null, undefined, false],
    [NaN, NaN, true],
    ['hi', 'hi', true],
    [5, 5, true],
    [5, 10, false],
    [[], [], true],
    [[1, 2], [1, 2], true],
    [[1, 2], [2, 1], false],
    [[1, 2], [1, 2, 3], false],
    [{}, {}, true],
    [{ a: 1, b: 2 }, { a: 1, b: 2 }, true],
    [{ a: 1, b: 2 }, { b: 2, a: 1 }, true],
    [{ a: 1, b: 2 }, { a: 1, b: 3 }, false],
    [{ a: 1, b: 2 }, { a: 1, b: 2, c: 3 }, false],
    [
      { 1: { name: 'foo', age: 28 }, 2: { name: 'bar', age: 26 } },
      { 1: { name: 'foo', age: 28 }, 2: { name: 'bar', age: 26 } },
      true,
    ],
    [
      { 1: { name: 'foo', age: 28 }, 2: { name: 'bar', age: 26 } },
      { 1: { name: 'foo', age: 28 }, 2: { name: 'bar', age: 27 } },
      false,
    ],
    [
      function (x: unknown) {
        return x
      },
      function (x: unknown) {
        return x
      },
      true,
    ],
    [
      function (x: unknown) {
        return x
      },
      function (y: unknown) {
        return (y as number) + 2
      },
      false,
    ],
  ]
  for (const [a, b, rv] of inputs) {
    expect(deepCompare(a, b)).toBe(rv)
  }
})
