import { ComponentProps, ReactNode, RefObject, memo, useCallback, useMemo, useRef } from 'react'
import { Virtuoso as VirtualList, VirtuosoHandle } from 'react-virtuoso'

import { toSorted } from '@/utils/array.ts'
import { cx } from '@/utils/css.ts'
import { useUpdateEffect } from '@/utils/hooks/use-update-effect.ts'

import { Image } from '@/components/image-grid.tsx'
import { T } from '@/components/translation.tsx'

import { ListMutator } from '@/views/actions.ts'

import { useInfinitePosts } from './hooks/use-infinite.ts'
import { PostCard } from './post-card.tsx'

export interface Post {
  id: number
  content: string
  files?: Image[]
  color: PostColor | null
  created_at: number
  updated_at: number
  deleted_at?: number
  shared: boolean
  parent?: Post
  parent_id?: number | null
  children_count: number
  score?: number
  tags?: string[]
}

export interface PostPagination {
  posts: Post[]
  cursor: number
  size: number
}

export type PostColor = 'red' | 'green' | 'blue'

export type OrderBy = 'created_at' | 'updated_at' | 'deleted_at' | 'score'

interface PostListProps extends ComponentProps<typeof VirtualList> {
  queryString?: string
  useWindowScroll?: boolean
  scrollParent?: HTMLElement
  showPlaceholder?: boolean
  orderBy?: OrderBy
  ascending?: boolean
  mutateRef?: RefObject<ListMutator>
}

export const PostList = memo(function PostList({
  queryString,
  scrollParent,
  useWindowScroll = false,
  showPlaceholder = true,
  orderBy,
  ascending = false,
  mutateRef,
  className,
  ...props
}: PostListProps) {
  const { posts, isEmpty, isLoadingMore, isReachingEnd, mutate, setSize } =
    useInfinitePosts(queryString)

  if (mutateRef) mutateRef.current = mutate

  const filteredAndSortedPosts = useMemo(() => {
    const isHiddenPage = queryString?.includes('tag=hidden')
    const filteredPosts = isHiddenPage
      ? posts
      : posts.filter((post) => !(post.tags ?? []).some((tag) => tag.startsWith('hidden')))

    if (!orderBy) return filteredPosts
    const key = orderBy
    return toSorted(filteredPosts, (x, y) => (ascending ? x[key]! - y[key]! : y[key]! - x[key]!))
  }, [queryString, posts, orderBy, ascending])

  const listHandle = useRef<VirtuosoHandle>(null!)

  useUpdateEffect(() => {
    listHandle.current.scrollToIndex(0)
  }, [queryString, orderBy, ascending])

  const scrollItemIntoView = useCallback((index: number) => {
    listHandle.current.scrollIntoView({ index })
  }, [])

  const isMainList = !queryString?.includes('parent_id')
  const isRecyclerPage = !!queryString?.includes('deleted=true')

  let footer: ReactNode = null
  if (isLoadingMore)
    footer = (
      <Footer>
        <T name="loading" />
      </Footer>
    )
  else if (isEmpty)
    footer = showPlaceholder ? (
      <Footer>
        <T name="noContent" />
      </Footer>
    ) : null
  else if (isReachingEnd) footer = <Footer>✨</Footer>

  return (
    <VirtualList
      ref={listHandle}
      className={className}
      useWindowScroll={useWindowScroll}
      customScrollParent={scrollParent}
      data={filteredAndSortedPosts}
      endReached={() => {
        if (isLoadingMore || isReachingEnd) return
        void setSize((size) => size + 1)
      }}
      itemContent={(index, post) => (
        <PostCard
          className={cx({ 'pt-4': index !== 0 })}
          key={(post as Post).id}
          index={index}
          scrollIntoView={scrollItemIntoView}
          collapsible
          showParentLink={isMainList && !isRecyclerPage}
          post={post as Post} // NOTE: something wrong with TS definition of Virtuoso
          mutator={mutate}
        />
      )}
      increaseViewportBy={200}
      components={{ Footer: () => footer }}
      {...props}
    />
  )
})

function Footer({ children }: ComponentProps<'div'>) {
  return <div className="text-muted-foreground/80 mt-2 mb-1 text-center text-sm">{children}</div>
}
