import { Editor, Node, Point, Range, Transforms } from 'slate'
import { RenderElementProps } from 'slate-react'

import { BLOCK_QUOTE, PARAGRAPH } from '../types'
import { findElement, isElementActive } from '../utils'
import { insertNewParagraph } from './paragraph'

export function withBlockQuote(editor: Editor) {
  const { insertBreak, insertText } = editor

  // When pressing `enter` in a block-quote:
  // 1. If not at the end of the line, insert `\n`
  // 2. If at the end of the line, and the previous character is not `\n`, insert `\n`
  // 3. If at the end of the line, and the previous character is `\n`, delete `\n` and insert a new paragraph
  editor.insertBreak = () => {
    const { selection } = editor

    if (selection && Range.isCollapsed(selection)) {
      const entry = findElement(editor, BLOCK_QUOTE)

      if (entry) {
        const [node, path] = entry
        const endPoint = Editor.end(editor, path)

        // If the cursor is at the end of a block-quote
        if (Point.equals(endPoint, selection.focus)) {
          const content = Node.string(node)
          // If the last character is `\n`
          if (content.endsWith('\n')) {
            // Then delete it
            editor.deleteBackward('character')
            // And insert a paragraph
            insertNewParagraph(editor)
            return
          }
        }

        insertText('\n')
        return
      }
    }

    insertBreak()
  }

  return editor
}

export function toggleBlockQuote(editor: Editor) {
  // Block-quote can only switch between paragraphs
  if (!isElementActive(editor, [BLOCK_QUOTE, PARAGRAPH])) {
    return
  }

  const active = isElementActive(editor, BLOCK_QUOTE)
  const newType = active ? PARAGRAPH : BLOCK_QUOTE

  Transforms.setNodes(editor, { type: newType }, { mode: 'lowest' })
}

export function BlockQuote({ attributes, children }: RenderElementProps) {
  return <blockquote {...attributes}>{children}</blockquote>
}
