import { Editor, Element, Range, Transforms } from 'slate'
import { ReactEditor, RenderElementProps, useReadOnly, useSlateStatic } from 'slate-react'

import { CHECK_LIST, CheckListElement, PARAGRAPH } from '../types'
import { findElement, isElementActive, isEmptyLine } from '../utils'

export function withCheckList(editor: Editor) {
  const { insertBreak } = editor

  editor.insertBreak = () => {
    const { selection } = editor

    if (selection && Range.isCollapsed(selection)) {
      const entry = findElement(editor, CHECK_LIST)
      if (entry) {
        const path = entry[1]
        if (isEmptyLine(editor, path)) {
          Transforms.setNodes(
            editor,
            { type: PARAGRAPH },
            {
              at: path,
            },
          )
        } else {
          // NOTE: Override the behavior of `insertBreak` of `ul` or `ol`
          Transforms.splitNodes(editor, {
            match: (node) => Element.isElement(node) && node.type === CHECK_LIST,
            mode: 'lowest',
            // If `true`, split at the beginning and end of the line as well
            always: true,
          })
        }
        return
      }
    }

    insertBreak()
  }

  return editor
}

export function toggleCheckList(editor: Editor) {
  // Check-list can only switch between paragraphs
  if (!isElementActive(editor, [CHECK_LIST, PARAGRAPH])) {
    return
  }

  const active = isElementActive(editor, CHECK_LIST)
  const newType = active ? PARAGRAPH : CHECK_LIST

  Transforms.setNodes(editor, { type: newType, checked: false } as Partial<CheckListElement>, {
    mode: 'lowest',
  })
}

export function CheckList({ attributes, children, element }: RenderElementProps) {
  const editor = useSlateStatic()
  const readOnly = useReadOnly()
  const { checked } = element as CheckListElement

  return (
    <div className="check-list" {...attributes}>
      <input
        contentEditable={false}
        type="checkbox"
        checked={checked}
        onChange={(event) => {
          const path = ReactEditor.findPath(editor as ReactEditor, element)
          Transforms.setNodes(
            editor,
            { checked: event.target.checked } as Partial<CheckListElement>,
            {
              at: path,
            },
          )
        }}
      />
      <label contentEditable={!readOnly} suppressContentEditableWarning>
        {children}
      </label>
    </div>
  )
}
