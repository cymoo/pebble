import { ComponentProps } from 'react'
import { Location, useLocation } from 'react-router'

import { cx } from '@/utils/css.ts'

import { Button } from '@/components/button.tsx'
import { useStableNavigate } from '@/components/router.tsx'

import { Post } from './post-list.tsx'

interface TruncateLinkProps extends ComponentProps<typeof Button> {
  post: Post
  maxLength?: number
}

export function TruncateLink({ post, maxLength, className, ...props }: TruncateLinkProps) {
  // NOTE: `useNavigate` hook causes waste rendering
  // https://github.com/remix-run/react-router/issues/7634
  const navigate = useStableNavigate()
  const location = useLocation()

  return (
    <Button
      className={cx('h-8', className)}
      variant="link"
      onClick={() => {
        const bg = (location.state as { backgroundLocation?: Location } | null)?.backgroundLocation
        void navigate(`/p/${String(post.id)}`, {
          state: {
            post,
            isFirstLayer: !bg,
            backgroundLocation: bg || location,
          },
        })
      }}
      {...props}
    >
      <span className="max-w-full truncate">
        {post.content.replace(/(<([^>]+)>)/gi, ' ').substring(0, maxLength)}
      </span>
    </Button>
  )
}
