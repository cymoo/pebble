import { ComponentProps, useRef, useState } from 'react'
import { Navigate, useLocation, useParams } from 'react-router'
import useSWR from 'swr'

import { cx } from '@/utils/css.ts'
import { isInteger } from '@/utils/text.ts'

import { Button } from '@/components/button.tsx'
import { useModal } from '@/components/modal.tsx'
import { Spinner } from '@/components/spinner.tsx'
import { T } from '@/components/translation.tsx'

import { ListMutator } from '@/views/actions.ts'
import { PostEditor } from '@/views/editor/editor.tsx'
import { CenteredContainer } from '@/views/layout/layout.tsx'

import { GET_POST, fetcher } from '@/api.ts'
import { AppError } from '@/error.ts'

import { PostCard } from './post-card.tsx'
import { Post, PostList } from './post-list.tsx'

export function PostPage({ className, ...props }: ComponentProps<'div'>) {
  const location = useLocation()
  const { id } = useParams() as { id: string }

  const {
    data: post,
    mutate,
    error,
  } = useSWR<Post, AppError>(
    [GET_POST, id],
    ([url, id]: [string, string]) => fetcher(`${url}?id=${id}`),
    { revalidateOnMount: true, fallbackData: (location.state as { post: Post } | null)?.post },
  )

  const [scrollElement, setScrollElement] = useState<HTMLElement>(null!)

  const listMutateRef = useRef<ListMutator>(null!)

  const modal = useModal()

  const hasBackground = !!(location.state as { backgroundLocation?: Location } | null)
    ?.backgroundLocation

  if (!isInteger(id) || error?.code === 404) {
    return <Navigate to="/404" replace />
  }

  if (!post) {
    return (
      <CenteredContainer title="loading...">
        <Spinner className="text-primary" />
      </CenteredContainer>
    )
  }

  return (
    <div
      ref={(ref) => {
        if (ref) setScrollElement(ref)
      }}
      className={cx({ 'vh-full overflow-y-auto': hasBackground }, className)}
      {...props}
    >
      <PostCard
        className="*:border-0 *:bg-transparent *:shadow-none"
        standalone
        post={post}
        mutator={mutate}
      />
      <DotSeparator className="py-3" />
      <div className="px-4">
        <Button
          className="w-full my-3"
          variant="outline"
          onClick={() => {
            modal.open({
              content: (
                <PostEditor
                  className="min-h-[150px]"
                  post={{ parent_id: post.id }}
                  mutator={listMutateRef.current}
                  afterSubmit={() => {
                    modal.close()
                  }}
                  afterCancel={() => {
                    modal.close()
                  }}
                />
              ),
            })
          }}
        >
          <T name="addMemo" />
        </Button>
        <PostList
          useWindowScroll={!hasBackground}
          scrollParent={hasBackground ? scrollElement : undefined}
          mutateRef={listMutateRef}
          queryString={`parent_id=${String(id)}`}
          showPlaceholder={false}
        />
      </div>
    </div>
  )
}

function DotSeparator({ className }: ComponentProps<'div'>) {
  return (
    <div className={cx('flex gap-5 justify-center', className)}>
      <div className="size-[3px] rounded-full bg-muted-foreground"></div>
      <div className="size-[3px] rounded-full bg-muted-foreground"></div>
      <div className="size-[3px] rounded-full bg-muted-foreground"></div>
    </div>
  )
}
