import { Tag, TagNode } from './tag-list.tsx'

export function buildTagTree(tags: Tag[]): TagNode[] {
  const nodeMap = new Map<string, TagNode>()

  for (const tag of tags) {
    nodeMap.set(tag.name, { ...tag, children: [] })
  }

  for (const tag of tags) {
    for (const parentName of getParentPaths(tag.name)) {
      if (!nodeMap.has(parentName)) {
        nodeMap.set(parentName, {
          name: parentName,
          sticky: false,
          post_count: -1,
          children: [],
        })
      }
    }
  }

  // Build tree structure by connecting children to their parent nodes
  for (const name of nodeMap.keys()) {
    const node = nodeMap.get(name)!
    const parentName = getParentPath(node.name)
    if (parentName) {
      const parentNode = nodeMap.get(parentName)!
      parentNode.children.push(node)
    }
  }

  function updateCountAndSort(nodes: TagNode[]): TagNode[] {
    for (const node of nodes) {
      // First recursively update all children
      node.children = updateCountAndSort(node.children)

      if (node.post_count === -1) {
        node.post_count = node.children.reduce((count, child) => count + child.post_count, 0)
      }
    }

    return nodes.filter((node) => node.post_count > 0).sort((a, b) => a.name.localeCompare(b.name))
  }

  const rootNodes = Array.from(nodeMap.values()).filter((node) => !node.name.includes('/'))
  return updateCountAndSort(rootNodes)
}

export function getStickyTags(tags: Tag[]): TagNode[] {
  return tags
    .map((tag) => ({ ...tag, children: [] }))
    .filter((tag) => tag.sticky && tag.post_count > 0)
    .sort((a, b) => a.name.localeCompare(b.name))
}

export function getLastSegment(str: string): string {
  const segments = str.split('/')
  return segments[segments.length - 1]
}

function getParentPaths(name: string): string[] {
  const paths: string[] = []
  let currentPath = ''

  for (const part of name.split('/').slice(0, -1)) {
    currentPath = currentPath ? `${currentPath}/${part}` : part
    paths.push(currentPath)
  }
  return paths
}

function getParentPath(path: string): string {
  const parts = path.split('/')
  if (parts.length <= 1) return ''
  return parts.slice(0, -1).join('/')
}
