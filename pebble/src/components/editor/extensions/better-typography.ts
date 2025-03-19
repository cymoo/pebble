import { Editor, Range } from 'slate'

import { CODE_BLOCK } from '../types'
import { getPrevChar } from '../utils'
import { isElementActive } from '../utils'

// Refer to tiptap for more replacement rules.
export function withBetterTypography(editor: Editor): Editor {
  const { insertText, deleteBackward } = editor

  editor.insertText = (text) => {
    // Characters should not be replaced within code blocks.
    if (isElementActive(editor, CODE_BLOCK)) {
      insertText(text)
      return
    }

    const pc = getPrevChar(editor)

    if (pc === '-' && text === '-') {
      Editor.deleteBackward(editor, { unit: 'character' })
      Editor.insertText(editor, '—')
      return
    }
    if (pc === '<' && text === '-') {
      Editor.deleteBackward(editor, { unit: 'character' })
      Editor.insertText(editor, '←')
      return
    }
    if (pc === '-' && text === '>') {
      Editor.deleteBackward(editor, { unit: 'character' })
      Editor.insertText(editor, '→')
      return
    }
    if (pc === '!' && text === '=') {
      Editor.deleteBackward(editor, { unit: 'character' })
      Editor.insertText(editor, '≠')
      return
    }

    insertText(text)
  }

  editor.deleteBackward = (...args) => {
    const { selection } = editor
    const pc = getPrevChar(editor)

    deleteBackward(...args)

    if (selection && Range.isCollapsed(selection)) {
      if (pc === '—') {
        Editor.insertText(editor, '--')
      }
      if (pc === '←') {
        Editor.insertText(editor, '<-')
      }
      if (pc === '→') {
        Editor.insertText(editor, '->')
      }
      if (pc === '≠') {
        Editor.insertText(editor, '!=')
      }
    }
  }

  return editor
}
