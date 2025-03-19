import { Tag, TagNode } from './tag-list.tsx'

export function getTreeTags(tags: Tag[]): TagNode[] {
  const nodeMap = new Map<string, TagNode>()

  // Initialize all nodes in nodeMap
  for (const item of tags) {
    nodeMap.set(item.name, { ...item, children: [] })
  }

  const rootNodes: TagNode[] = []

  // Build tree structure by connecting children to their parent nodes
  for (const { name } of tags) {
    const node = nodeMap.get(name)!
    const parts = name.split('/')

    if (parts.length > 1) {
      const parentName = parts.slice(0, -1).join('/')
      const parentNode = nodeMap.get(parentName)!
      if (node.post_count > 0) parentNode.children.push(node)
    } else {
      if (node.post_count > 0) rootNodes.push(node)
    }
  }

  function sort(tags: TagNode[]): TagNode[] {
    tags.sort((a, b) => a.name.localeCompare(b.name))

    tags.forEach((tag) => {
      if (tag.children.length > 0) {
        tag.children = sort(tag.children)
      }
    })

    return tags
  }

  return sort(rootNodes)
}

export function getStickyTags(tags: Tag[]): TagNode[] {
  return tags
    .map((tag) => ({ ...tag, children: [] }))
    .filter((tag) => tag.sticky && tag.post_count > 0)
    .sort((a, b) => a.name.localeCompare(b.name))
}

export function extractLastSegment(str: string): string {
  const segments = str.split('/')
  return segments[segments.length - 1]
}
