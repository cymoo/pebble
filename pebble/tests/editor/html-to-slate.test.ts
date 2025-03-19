import { fromHtml } from '../../src/components/editor/html'

test('convert html string to slate nodes', () => {
  const inputs: Array<[string, object[]]> = [
    ['<p>foo</p>', [{ type: 'paragraph', children: [{ text: 'foo' }] }]],

    ['<p>foo</p>bar', [{ type: 'paragraph', children: [{ text: 'foo' }] }, { text: 'bar' }]],

    [
      '<p>foo</p><span>bar</span>',
      [{ type: 'paragraph', children: [{ text: 'foo' }] }, { text: 'bar' }],
    ],

    [
      '<div><p>foo</p><span>bar</span></div>',
      [{ type: 'paragraph', children: [{ text: 'foo' }] }, { text: 'bar' }],
    ],

    ['<div><p>foo</p></div>', [{ type: 'paragraph', children: [{ text: 'foo' }] }]],

    [
      '<p><i><u>foo</u></i></p>',
      [{ type: 'paragraph', children: [{ text: 'foo', italic: true, underline: true }] }],
    ],

    ['<ul><p>foo</p></ul>', [{ text: '' }]],

    [
      '<ul><p>foo</p><li>bar</li></ul>',
      [
        {
          type: 'bulleted-list',
          children: [
            { type: 'list-item', children: [{ type: 'paragraph', children: [{ text: 'bar' }] }] },
          ],
        },
      ],
    ],
  ]

  for (const [input, output] of inputs) {
    const node = new DOMParser().parseFromString(input, 'text/html').body
    const result = fromHtml(node)
    expect(result).toEqual(output)
  }
})
