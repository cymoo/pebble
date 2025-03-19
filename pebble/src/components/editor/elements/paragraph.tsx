import { Editor, Location, Transforms } from 'slate'
import { RenderElementProps } from 'slate-react'

import { PARAGRAPH, ParagraphElement } from '../types'

export function Paragraph({ attributes, children }: RenderElementProps) {
  return <p {...attributes}>{children}</p>
}

export function insertNewParagraph(editor: Editor, at?: Location) {
  Transforms.insertNodes(
    editor,
    {
      type: PARAGRAPH,
      children: [{ text: '' }],
    } as ParagraphElement,
    {
      mode: 'lowest',
      at,
    },
  )
}
