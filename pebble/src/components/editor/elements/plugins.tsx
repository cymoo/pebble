import { Editor, Point, Range, Transforms } from 'slate'

import { PARAGRAPH } from '../types'
import { findElements } from '../utils'
import { insertNewParagraph } from './paragraph'

export function withAvoidEmptyChildren(editor: Editor): Editor {
  const { normalizeNode } = editor

  editor.normalizeNode = (entry) => {
    const [, path] = entry
    if (path.length === 0) {
      if (editor.children.length === 0) {
        insertNewParagraph(editor, [0])
      }
    }

    normalizeNode(entry)
  }

  return editor
}

export function withResetToParagraphWhenDeleteAtBlockStart(...types: string[]) {
  return (editor: Editor) => {
    const { deleteBackward } = editor

    editor.deleteBackward = (...args) => {
      const { selection } = editor

      if (selection && Range.isCollapsed(selection)) {
        const [entry] = findElements(editor, types)
        if (typeof entry !== 'undefined') {
          const [, path] = entry
          const startPoint = Editor.start(editor, path)

          if (Point.equals(startPoint, selection.focus)) {
            Transforms.setNodes(
              editor,
              { type: PARAGRAPH },
              {
                at: path,
              },
            )
            return
          }
        }
      }

      deleteBackward(...args)
    }

    return editor
  }
}
