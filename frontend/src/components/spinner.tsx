import { ComponentProps } from 'react'

import { cx } from '@/utils/css.ts'

const sizes = {
  xs: 'size-4',
  sm: 'size-5',
  md: 'size-6',
  lg: 'size-8',
  xl: 'size-10',
}

interface SpinnerProps extends ComponentProps<'span'> {
  size?: keyof typeof sizes
}

export function Spinner({ size = 'md', className, ...props }: SpinnerProps) {
  return (
    <span
      className={cx('pointer-events-none inline-block', sizes[size], className)}
      role="status"
      {...props}
    >
      <span className="inline-block h-full w-full animate-spin rounded-full border border-transparent border-t-current! border-l-current! [animation-duration:750ms]">
        <span className="sr-only">Loading...</span>
      </span>
    </span>
  )
}
