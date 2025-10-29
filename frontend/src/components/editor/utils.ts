import {
  Ancestor,
  Descendant,
  Editor,
  Element,
  Location,
  Node,
  NodeEntry,
  Path,
  Point,
  Range,
  Transforms,
} from 'slate'
import { ReactEditor } from 'slate-react'

import { toggleBlockQuote } from './elements/block-quote'
import { toggleCheckList } from './elements/check-list'
import { toggleCodeBlock } from './elements/code-block'
import { toggleHeading } from './elements/heading'
import { toggleList } from './elements/list'
import {
  BLOCK_QUOTE,
  BULLETED_LIST,
  CHECK_LIST,
  CODE_BLOCK,
  MarkType,
  NUMBERED_LIST,
  PARAGRAPH,
  isHeading,
} from './types'

export function isBlockStart(editor: Editor): boolean {
  const { selection } = editor
  if (selection && Range.isCollapsed(selection)) {
    const match = Editor.above(editor, {
      match: (node) =>
        Element.isElement(node) && Editor.isBlock(editor, node) && !Editor.isVoid(editor, node),
    })

    if (match) {
      const [, path] = match
      const start = Editor.start(editor, path)

      if (Point.equals(selection.anchor, start)) {
        return true
      }
    }
  }

  return false
}

export function isEditorEmpty(editor: Editor): boolean {
  if (editor.children.length === 0) {
    return true
  }

  if (
    editor.children.length === 1 &&
    !editor.isVoid(editor.children[0] as Element) &&
    Node.string(editor.children[0]) === ''
  ) {
    for (const [node] of Node.children(editor, [0])) {
      if (Element.isElement(node) && editor.isVoid(node)) {
        return false
      }
    }
    return true
  }

  return false
}

export function isFirstElementParagraph(editor: Editor): boolean {
  return (
    editor.children.length !== 0 &&
    Element.isElement(editor.children[0]) &&
    editor.children[0].type === PARAGRAPH
  )
}

export function getCurrentBlock(editor: Editor): NodeEntry<Ancestor> | undefined {
  return Editor.above(editor, {
    match: (node) =>
      Element.isElement(node) && Editor.isBlock(editor, node) && !Editor.isVoid(editor, node),
    mode: 'lowest',
  })
}

export function getPrevChar(editor: Editor, point?: Location): string {
  if (point === undefined) {
    if (!editor.selection) {
      throw Error('cannot get the previous character without a selection')
    }
    point = editor.selection.anchor
  }

  // NOTE: use `Editor.before(editor, point, {unit: 'character'})`?
  const prevPoint = Editor.before(editor, point)
  if (!prevPoint) return ''

  const range = Editor.range(editor, prevPoint, point)
  return Editor.string(editor, range)
}

export function getNextChar(editor: Editor, point?: Location): string {
  if (point === undefined) {
    if (!editor.selection) {
      throw Error('cannot get the previous character without a selection')
    }
    point = editor.selection.focus
  }

  // NOTE: use `Editor.after(editor, point, {unit: 'character'})`?
  const nextPoint = Editor.after(editor, point)
  if (!nextPoint) return ''

  const range = Editor.range(editor, point, nextPoint)
  return Editor.string(editor, range)
}

export function getFirstChild(editor: Editor, parentPath: Path): NodeEntry<Descendant> {
  const [firstChild] = Node.children(editor, parentPath)
  return firstChild
}

export function getLastChild(editor: Editor, parentPath: Path): NodeEntry<Descendant> {
  const [lastChild] = Node.children(editor, parentPath, {
    reverse: true,
  })
  return lastChild
}

export function isFirstNode(editor: Editor, path: Path): boolean {
  const prevNode = Editor.previous(editor, { at: path })
  return !prevNode
}

export function isLastNode(editor: Editor, path: Path): boolean {
  const nextNode = Editor.next(editor, { at: path })
  return !nextNode
}

export function isEmptyLine(editor: Editor, path: Path): boolean {
  return Point.equals(Editor.start(editor, path), Editor.end(editor, path))
}

export function isElementActive(
  editor: Editor,
  type: string | string[],
  universal = true,
): boolean {
  const { selection } = editor
  if (selection === null) {
    return false
  }

  const [entry] = Editor.nodes(editor, {
    match: (node) =>
      Element.isElement(node) &&
      (Array.isArray(type) ? type.includes(node.type) : node.type === type),
    at: Editor.unhangRange(editor, selection),
    mode: 'lowest',
    // If true, all elements in the selection need to be active.
    universal,
  })

  return !!entry
}

export function findElements(
  editor: Editor,
  types: string[] | string,
  universal = false,
  at?: Location,
) {
  if (!Array.isArray(types)) {
    types = [types]
  }
  return Editor.nodes(editor, {
    match: (node) => Element.isElement(node) && types.includes(node.type),
    at: at ?? Editor.unhangRange(editor, editor.selection!),
    mode: 'lowest',
    universal,
  })
}

export function findElement(
  editor: Editor,
  type: string,
  universal = false,
  at?: Location,
): NodeEntry | null {
  const [entry] = findElements(editor, type, universal, at)
  if (typeof entry === 'undefined') {
    return null
  }
  return entry
}

export function toggleBlock(editor: Editor, format: string) {
  if (editor.selection === null) {
    return
  }
  // fixChromeDoubleClickBug(editor)

  if (isHeading(format)) {
    toggleHeading(editor, format)
    return
  }

  if (format === BLOCK_QUOTE) {
    toggleBlockQuote(editor)
    return
  }

  if (format === CODE_BLOCK) {
    toggleCodeBlock(editor)
    return
  }

  if (format === CHECK_LIST) {
    toggleCheckList(editor)
    return
  }

  if (format === NUMBERED_LIST || format === BULLETED_LIST) {
    toggleList(editor, format)
    return
  }
}

export function liftNodeAndUnWrap(editor: Editor, path: Path) {
  Editor.withoutNormalizing(editor, () => {
    const pathRef = Editor.pathRef(editor, path)

    Transforms.liftNodes(editor, {
      at: pathRef.current!,
    })

    Transforms.unwrapNodes(editor, {
      at: pathRef.current!,
    })

    pathRef.unref()
  })
}

export function isMarkActive(editor: Editor, format: MarkType): boolean {
  const marks = Editor.marks(editor)
  return marks ? marks[format] === true : false
}

export function toggleMark(editor: Editor, format: MarkType) {
  if (editor.selection === null) {
    return
  }

  // Adding marks is not allowed within a code-block.
  if (isElementActive(editor, CODE_BLOCK)) {
    return
  }

  const isActive = isMarkActive(editor, format)

  if (isActive) {
    Editor.removeMark(editor, format)
  } else {
    Editor.addMark(editor, format, true)
  }
}

export function removeAllMarks(editor: Editor) {
  const marks = Editor.marks(editor)
  for (const mark in marks) {
    Editor.removeMark(editor, mark)
  }
}

export function setInitialContent(editor: ReactEditor, content: Descendant[]) {
  Editor.withoutNormalizing(editor, () => {
    Transforms.removeNodes(editor, { at: [0] })
    Transforms.insertNodes(editor, content, { at: [0] })
  })
}

export const isUrl = (text: string) => /^https?:\/\/\S+$/.test(text)

export const imageExtensions = ['gif', 'jpg', 'jpeg', 'png', 'svg', 'webp']
export const isImageUrl = (url: string) => {
  if (!url) {
    return false
  }
  if (!isUrl(url)) {
    return false
  }

  const ext = new URL(url).pathname.split('.').pop()
  return !!imageExtensions.find((item) => item == ext)
}

// NOTE: There is a bug in Chrome (~2020) where double-clicking to select text incorrectly selects the zero-width character in the next block void element.
// It is unclear whether recent versions of Chrome have fixed this issue.
export function fixChromeDoubleClickBug(editor: Editor) {
  const { selection } = editor
  if (selection && selection.focus.offset === 1) {
    const [node, _] = Editor.node(editor, Path.parent(selection.focus.path))
    if (Element.isElement(node) && Editor.isVoid(editor, node) && Editor.isBlock(editor, node)) {
      Transforms.move(editor, { reverse: true, edge: 'focus' })
    }
  }
}
