import { Editor, Range, Transforms } from 'slate'

import { getNextChar, getPrevChar } from '../utils'

export function withMatchBrackets(editor: Editor): Editor {
  const { insertText, deleteBackward } = editor

  editor.insertText = (text) => {
    const pc = getPrevChar(editor)
    const nc = getNextChar(editor)

    if (
      (text === ']' && pc === '[' && nc === ']') ||
      (text === ')' && pc === '(' && nc === ')') ||
      (text === '}' && pc === '{' && nc === '}') ||
      (text === '>' && pc === '<' && nc === '>') ||
      (text === '"' && pc === '"' && nc === '"') ||
      (text === "'" && pc === "'" && nc === "'")
    ) {
      Transforms.move(editor)
      return
    }

    insertText(text)

    if ('[({<'.includes(text)) {
      if (text === '[') Editor.insertText(editor, ']')
      if (text === '(') Editor.insertText(editor, ')')
      if (text === '{') Editor.insertText(editor, '}')
      if (text === '<') Editor.insertText(editor, '>')
      Transforms.move(editor, { reverse: true })
    }
    if (text === '"' && pc !== '"') {
      Editor.insertText(editor, '"')
      Transforms.move(editor, { reverse: true })
    }
    if (text === "'" && pc !== "'") {
      Editor.insertText(editor, "'")
      Transforms.move(editor, { reverse: true })
    }
  }

  editor.deleteBackward = (...args) => {
    const { selection } = editor

    if (selection && Range.isCollapsed(selection)) {
      const pc = getPrevChar(editor)
      const nc = getNextChar(editor)
      if (
        (pc === '[' && nc === ']') ||
        (pc === '(' && nc === ')') ||
        (pc === '{' && nc === '}') ||
        (pc === '<' && nc === '>') ||
        (pc === '"' && nc === '"') ||
        (pc === "'" && nc === "'")
      ) {
        Editor.deleteForward(editor, { unit: 'character' })
      }
    }

    deleteBackward(...args)
  }

  return editor
}
