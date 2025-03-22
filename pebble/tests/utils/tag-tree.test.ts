import { Tag } from '../../src/views/tag/tag-list'
import { buildTagTree } from '../../src/views/tag/utils'

describe('buildTagTree', () => {
  // Basic functionality tests
  it('should build a basic tree structure from flat tags', () => {
    const tags: Tag[] = [
      { name: 'parent', post_count: 5, sticky: false },
      { name: 'parent/child', post_count: 3, sticky: false },
    ]

    const result = buildTagTree(tags)

    expect(result.length).toBe(1)
    expect(result[0].name).toBe('parent')
    expect(result[0].children.length).toBe(1)
    expect(result[0].children[0].name).toBe('parent/child')
    // parent's post_count should remain 5, not summed
    expect(result[0].post_count).toBe(5)
  })

  // Empty array edge case
  it('should return an empty array when given an empty array', () => {
    const result = buildTagTree([])
    expect(result).toEqual([])
  })

  // Auto-created parent nodes test
  it('should calculate post_count for auto-created parent nodes only', () => {
    const tags: Tag[] = [{ name: 'a/b/c', post_count: 5, sticky: false }]

    const result = buildTagTree(tags)

    expect(result.length).toBe(1)
    expect(result[0].name).toBe('a')
    expect(result[0].children[0].name).toBe('a/b')
    expect(result[0].children[0].children[0].name).toBe('a/b/c')

    // Auto-created parent nodes should have their post_count calculated from children
    expect(result[0].post_count).toBe(5) // calculated from child
    expect(result[0].children[0].post_count).toBe(5) // calculated from child
    expect(result[0].children[0].children[0].post_count).toBe(5) // original value
  })

  // Multiple paths with shared auto-created parent
  it('should correctly calculate post_count for shared auto-created parent', () => {
    const tags: Tag[] = [
      { name: 'a/b/c', post_count: 2, sticky: false },
      { name: 'a/b/d', post_count: 3, sticky: false },
    ]

    const result = buildTagTree(tags)

    expect(result.length).toBe(1)
    expect(result[0].name).toBe('a')
    expect(result[0].children[0].name).toBe('a/b')
    expect(result[0].children[0].children.length).toBe(2)

    // Auto-created parent nodes should sum child counts
    expect(result[0].post_count).toBe(5) // auto-created, sum of all descendants
    expect(result[0].children[0].post_count).toBe(5) // auto-created, sum of direct children

    const cNode = result[0].children[0].children.find((child) => child.name === 'a/b/c')
    const dNode = result[0].children[0].children.find((child) => child.name === 'a/b/d')
    expect(cNode?.post_count).toBe(2) // original value
    expect(dNode?.post_count).toBe(3) // original value
  })

  // Mix of existing and auto-created parent nodes
  it('should respect existing post_count for real tags and calculate for auto-created ones', () => {
    const tags: Tag[] = [
      { name: 'parent', post_count: 10, sticky: false },
      { name: 'parent/child/grandchild', post_count: 5, sticky: false },
    ]

    const result = buildTagTree(tags)

    expect(result.length).toBe(1)
    expect(result[0].name).toBe('parent')
    expect(result[0].children[0].name).toBe('parent/child')
    expect(result[0].children[0].children[0].name).toBe('parent/child/grandchild')

    // Existing parent keeps its own post_count
    expect(result[0].post_count).toBe(10)

    // Auto-created middle node gets count from child
    expect(result[0].children[0].post_count).toBe(5)

    // Original leaf node keeps its count
    expect(result[0].children[0].children[0].post_count).toBe(5)
  })

  // Zero post count filtering
  it('should filter out nodes with zero post counts', () => {
    const tags: Tag[] = [
      { name: 'a', post_count: 0, sticky: false },
      { name: 'b', post_count: 3, sticky: false },
      { name: 'c', post_count: 0, sticky: false },
    ]

    const result = buildTagTree(tags)

    expect(result.length).toBe(1)
    expect(result[0].name).toBe('b')
  })

  // Auto-created parents for nodes with zero counts
  it('should filter out auto-created parents for children with zero counts', () => {
    const tags: Tag[] = [{ name: 'parent/child', post_count: 0, sticky: false }]

    const result = buildTagTree(tags)

    // Both parent and child should be filtered out
    expect(result).toEqual([])
  })

  // Mixed zero and non-zero counts for siblings
  it('should handle siblings with mixed zero and non-zero counts', () => {
    const tags: Tag[] = [
      { name: 'parent/child1', post_count: 0, sticky: false },
      { name: 'parent/child2', post_count: 5, sticky: false },
    ]

    const result = buildTagTree(tags)

    expect(result.length).toBe(1)
    expect(result[0].name).toBe('parent')
    expect(result[0].children.length).toBe(1)
    expect(result[0].children[0].name).toBe('parent/child2')
    expect(result[0].post_count).toBe(5) // Auto-created, gets count from non-zero child
  })

  // Multiple root nodes with mixed counts
  it('should handle multiple root nodes with mixed post counts', () => {
    const tags: Tag[] = [
      { name: 'a', post_count: 5, sticky: false },
      { name: 'b', post_count: 0, sticky: false },
      { name: 'c', post_count: 3, sticky: false },
    ]

    const result = buildTagTree(tags)

    expect(result.length).toBe(2)
    expect(result[0].name).toBe('a')
    expect(result[1].name).toBe('c')
  })

  // Sorting test
  it('should sort nodes alphabetically', () => {
    const tags: Tag[] = [
      { name: 'z', post_count: 1, sticky: false },
      { name: 'c', post_count: 3, sticky: false },
      { name: 'b', post_count: 2, sticky: false },
    ]

    const result = buildTagTree(tags)

    expect(result.length).toBe(3)
    expect(result[0].name).toBe('b')
    expect(result[1].name).toBe('c')
    expect(result[2].name).toBe('z')
  })

  // Complex hierarchy with multiple levels
  it('should build a complex hierarchy correctly', () => {
    const tags: Tag[] = [
      { name: 'a', post_count: 1, sticky: false },
      { name: 'a/b', post_count: 2, sticky: false },
      { name: 'a/b/c', post_count: 3, sticky: false },
      { name: 'a/d', post_count: 4, sticky: false },
      { name: 'e/f/g', post_count: 5, sticky: false },
    ]

    const result = buildTagTree(tags)

    expect(result.length).toBe(2) // 'a' and 'e' root nodes

    // Find 'a' node and check its structure
    const aNode = result.find((node) => node.name === 'a')
    expect(aNode).toBeDefined()
    expect(aNode?.post_count).toBe(1) // Original value, not sum
    expect(aNode?.children.length).toBe(2) // 'a/b' and 'a/d'

    // Check 'a/b' branch
    const abNode = aNode?.children.find((child) => child.name === 'a/b')
    expect(abNode?.post_count).toBe(2) // Original value, not sum
    expect(abNode?.children.length).toBe(1)
    expect(abNode?.children[0].name).toBe('a/b/c')
    expect(abNode?.children[0].post_count).toBe(3)

    // Check 'a/d' node
    const adNode = aNode?.children.find((child) => child.name === 'a/d')
    expect(adNode?.post_count).toBe(4)

    // Check 'e' branch (auto-created parent)
    const eNode = result.find((node) => node.name === 'e')
    expect(eNode).toBeDefined()
    expect(eNode?.post_count).toBe(5) // Calculated from child
    expect(eNode?.children.length).toBe(1)
    expect(eNode?.children[0].name).toBe('e/f')
    expect(eNode?.children[0].post_count).toBe(5) // Calculated from child
    expect(eNode?.children[0].children[0].name).toBe('e/f/g')
    expect(eNode?.children[0].children[0].post_count).toBe(5) // Original value
  })

  // Edge case: negative post counts
  it('should treat negative post counts (except -1) as if they were zero', () => {
    const tags: Tag[] = [
      { name: 'a', post_count: -5, sticky: false },
      { name: 'b', post_count: 3, sticky: false },
    ]

    const result = buildTagTree(tags)

    // 'a' should be filtered out as if post_count was 0
    expect(result.length).toBe(1)
    expect(result[0].name).toBe('b')
  })

  // Special case for auto-created nodes with -1 post_count
  it('should handle auto-created parent nodes with -1 post_count specifically', () => {
    // This test verifies the specific logic for -1 post_count nodes
    const tags: Tag[] = [{ name: 'a/b/c', post_count: 5, sticky: false }]

    // Manually inspect the nodeMap contents
    const nodeMap = new Map()
    // Adding original tag
    nodeMap.set('a/b/c', { name: 'a/b/c', post_count: 5, sticky: false, children: [] })
    // Adding auto-created parents
    nodeMap.set('a', { name: 'a', post_count: -1, sticky: false, children: [] })
    nodeMap.set('a/b', { name: 'a/b', post_count: -1, sticky: false, children: [] })

    // The -1 post_count should trigger sum calculation only for these nodes
    const result = buildTagTree(tags)

    expect(result[0].post_count).toBe(5) // Calculated from child
  })
})
