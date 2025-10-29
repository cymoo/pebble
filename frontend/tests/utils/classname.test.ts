import { cx } from '../../src/utils/css'

test('classname', () => {
  let inputs = [
    [[null, undefined], ''],
    [['foo', undefined, 'bar'], 'foo bar'],
    [['foo', null, 'bar'], 'foo bar'],
    [['foo', { bar: true }], 'foo bar'],
    [['foo', { bar: false }], 'foo'],
    [['foo', { bar: true, fox: false }, 'box'], 'foo bar box'],
  ]
  for (const [input, rv] of inputs) {
    expect(cx(...input)).toEqual(rv)
  }
})
