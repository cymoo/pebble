import { chunk, range } from '../../src/utils/array'

test('chunk array', () => {
  expect(chunk(range(1, 6), 2)).toEqual([[1, 2], [3, 4], [5]])
  expect(chunk(range(1, 9), 3)).toEqual([
    [1, 2, 3],
    [4, 5, 6],
    [7, 8],
  ])
  expect(chunk(range(1, 4), 3)).toEqual([[1, 2, 3]])
})
