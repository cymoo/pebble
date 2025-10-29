import { Editor, Element, Point, Range, Transforms } from 'slate'
import { RenderElementProps } from 'slate-react'

import { HEADINGS, HeadingType, PARAGRAPH, isHeading } from '../types'
import { findElements, isElementActive } from '../utils'
import { insertNewParagraph } from './paragraph'

export function withHeading(editor: Editor): Editor {
  const { insertBreak, normalizeNode } = editor

  // Headings are only allowed at the top level; otherwise, convert them to paragraphs
  // NOTE:
  // 1. Only when the user types something into the editor, `normalizeNode` is called for that specific node and parent nodes
  // 2. Other nodes that have not been touched stay invalid until e.g. the text of that node is edited
  editor.normalizeNode = (entry) => {
    const [node, path] = entry

    if (Element.isElement(node) && isHeading(node.type)) {
      if (path.length > 1) {
        Transforms.setNodes(editor, { type: PARAGRAPH }, { at: path })
        return
      }
    }

    normalizeNode(entry)
  }

  // When pressing `enter` at the end of a heading line, insert a paragraph
  editor.insertBreak = () => {
    const { selection } = editor

    if (selection && Range.isCollapsed(selection)) {
      const [entry] = findElements(editor, HEADINGS as unknown as string[])

      if (typeof entry !== 'undefined') {
        const [, path] = entry
        const endPoint = Editor.end(editor, path)

        // If the cursor is at the end of a heading
        if (Point.equals(endPoint, selection.focus)) {
          insertNewParagraph(editor)
          return
        }
      }
    }

    insertBreak()
  }

  return editor
}

export function toggleHeading(editor: Editor, headingType: HeadingType) {
  // Headings can only switch between paragraphs
  if (!isElementActive(editor, [...HEADINGS, PARAGRAPH])) {
    return
  }

  const type = getHeadingType(editor)
  const newType = type === headingType ? PARAGRAPH : headingType

  Transforms.setNodes(editor, { type: newType }, { mode: 'lowest' })
}

function getHeadingType(editor: Editor): string | null {
  // Only return the heading type if all headings in the selection are of the same type;
  // otherwise, return null.
  const nodes = new Set(
    Array.from(findElements(editor, HEADINGS as unknown as string[], true)).map(
      (entry) => (entry[0] as Element).type,
    ),
  )
  return nodes.size === 1 ? [...nodes][0] : null
}

export function HeadingOne({ attributes, children }: RenderElementProps) {
  return <h1 {...attributes}>{children}</h1>
}
export function HeadingTwo({ attributes, children }: RenderElementProps) {
  return <h2 {...attributes}>{children}</h2>
}
export function HeadingThree({ attributes, children }: RenderElementProps) {
  return <h3 {...attributes}>{children}</h3>
}
export function HeadingFour({ attributes, children }: RenderElementProps) {
  return <h4 {...attributes}>{children}</h4>
}
export function HeadingFive({ attributes, children }: RenderElementProps) {
  return <h5 {...attributes}>{children}</h5>
}
