import { Editor, Element, Node, Point, Range, Transforms } from 'slate'
import { ReactEditor, RenderElementProps } from 'slate-react'

import { CODE_BLOCK, PARAGRAPH } from '../types'
import { findElement, isElementActive } from '../utils'
import { insertNewParagraph } from './paragraph'

export function withCodeBlock(editor: Editor): Editor {
  const reactEditor = editor as ReactEditor
  const { insertBreak, insertData } = reactEditor

  // When inside a code-block, read the clipboard content in plain text format
  reactEditor.insertData = (data) => {
    if (isElementActive(editor, CODE_BLOCK)) {
      const text = data.getData('text/plain')
      Editor.insertText(editor, text)
      return
    }

    insertData(data)
  }

  // When pressing `enter` in a code-block:
  // 1. If not at the end of the line, insert `\n`
  // 2. If at the end of the line, and the previous character is not `\n\n`, insert `\n`
  // 3. If at the end of the line, and the previous character is `\n\n`, delete `\n\n` and insert a new paragraph
  // 4. Preserve the number of leading spaces from the previous line
  editor.insertBreak = () => {
    const { selection } = editor

    if (selection && Range.isCollapsed(selection)) {
      const entry = findElement(editor, CODE_BLOCK)

      if (entry) {
        const [node, path] = entry
        const code = Node.string(node)

        const currentLine = getCurrentLine(editor, code)

        // Get the leading spaces of the current line
        let indent = ''
        const match = /^\s+/.exec(currentLine)
        if (match !== null) {
          indent = match[0]
        }

        const newLineWithIndent = `\n${indent}`
        const newLineWithIndentX2 = `${newLineWithIndent}${newLineWithIndent}`

        // If the cursor is at the end of a code-block and two consecutive enters have been typed
        const endPoint = Editor.end(editor, path)
        if (Point.equals(endPoint, selection.focus) && code.endsWith(newLineWithIndentX2)) {
          Transforms.delete(editor, {
            distance: newLineWithIndentX2.length,
            unit: 'character',
            reverse: true,
          })
          insertNewParagraph(editor)
          return
        } else {
          editor.insertText(newLineWithIndent)
          return
        }
      }
    }

    insertBreak()
  }

  return editor
}

export function toggleCodeBlock(editor: Editor) {
  // Block-quote can only switch between paragraphs
  if (!isElementActive(editor, [CODE_BLOCK, PARAGRAPH])) {
    return
  }

  const active = isElementActive(editor, CODE_BLOCK)

  if (active) {
    convertToParagraphs(editor)
  } else {
    convertToCodeBlock(editor)
  }
}

export function CodeBlock({ attributes, children }: RenderElementProps) {
  return (
    <pre {...attributes}>
      <code>{children}</code>
    </pre>
  )
}

// Get the content of the line where the cursor is located
function getCurrentLine(editor: Editor, code: string): string {
  const selection = editor.selection
  if (selection === null) {
    return ''
  }

  const offset = selection.anchor.offset
  const textBeforeCursor = code.slice(0, offset)
  const lines = textBeforeCursor.split('\n')
  const length = lines.length
  if (length === 0) {
    return ''
  }

  return lines[length - 1]
}

function convertToParagraphs(editor: Editor) {
  const nodes = Array.from(
    Editor.nodes(editor, {
      match: (node) => Element.isElement(node) && Editor.isBlock(editor, node),
      mode: 'lowest',
    }),
  )

  const code = nodes.map((entry) => Node.string(entry[0])).join('\n')

  const paragraphs = code
    .split('\n')
    .map((line) => ({ type: 'paragraph', children: [{ text: line }] }))

  // NOTE: `removeNodes` might remove the only node, in which case `normalize` will fail (children is empty),
  // so `withoutNormalizing` should be used.
  Editor.withoutNormalizing(editor, () => {
    Transforms.removeNodes(editor, {
      match: (node) => Element.isElement(node) && Editor.isBlock(editor, node),
      mode: 'lowest',
    })
    Transforms.insertNodes(editor, paragraphs, { mode: 'lowest' })
  })
}

function convertToCodeBlock(editor: Editor) {
  const nodes = Array.from(
    Editor.nodes(editor, {
      match: (node) => Element.isElement(node) && Editor.isBlock(editor, node),
      mode: 'lowest',
    }),
  )

  const code = nodes.map((entry) => Node.string(entry[0])).join('\n')

  Editor.withoutNormalizing(editor, () => {
    Transforms.removeNodes(editor, {
      match: (node) => Element.isElement(node) && Editor.isBlock(editor, node),
      mode: 'lowest',
    })
    Transforms.insertNodes(
      editor,
      { type: CODE_BLOCK, children: [{ text: code }] },
      { mode: 'lowest' },
    )
  })
}
