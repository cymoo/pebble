import { ComponentProps, memo, useEffect, useMemo } from 'react'
import useSWR from 'swr'

import { cx } from '@/utils/css.ts'

import { T } from '@/components/translation.tsx'

import { tagActions as actions } from '@/views/actions.ts'

import { GET_TAGS } from '@/api.ts'

import { TagItem } from './tag-item.tsx'
import { buildTagTree, getStickyTags } from './utils.ts'

export interface Tag {
  name: string
  post_count: number
  sticky: boolean
}

export interface TagNode extends Tag {
  children: TagNode[]
}

export const TagList = (props: ComponentProps<'div'>) => {
  const { data: tags, mutate } = useSWR<Tag[]>(GET_TAGS, { fallbackData: [] })

  const treeTags = useMemo(() => buildTagTree(tags!), [tags])
  const stickyTags = useMemo(() => getStickyTags(tags!), [tags])

  useEffect(() => {
    actions.tagsMutator = mutate
    return () => {
      actions.tagsMutator = undefined
    }
  }, [mutate])

  return (
    <div {...props}>
      {stickyTags.length > 0 && (
        <>
          <h3 className="flex items-center font-semibold">
            <T name="pinnedTags" />
          </h3>
          <div className="-mx-4 mt-2 mb-4 *:ml-0">
            <Tree treeData={stickyTags} showPath />
          </div>
        </>
      )}

      <h3 className="flex items-center font-semibold">
        <T name="allTags" />
      </h3>
      <div className="-mx-4 mt-2 mb-4 *:ml-0">
        <Tree treeData={treeTags} />
      </div>
    </div>
  )
}

interface TreeProps extends ComponentProps<'ul'> {
  treeData: TagNode[]
  showPath?: boolean
}

const Tree = memo(function Tree({ treeData, showPath = false, className, ...rest }: TreeProps) {
  if (treeData.length === 0) return null

  return (
    <ul className={cx('ml-3', className)} {...rest}>
      {treeData.map((item) => (
        <TagItem tag={item} key={item.name} showPath={showPath}>
          {item.children.length > 0 && <Tree treeData={item.children} />}
        </TagItem>
      ))}
    </ul>
  )
})
