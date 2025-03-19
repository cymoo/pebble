import { ComponentProps } from 'react'

import { cx } from '@/utils/css.ts'

const COLORS = { red: 'bg-red-500', blue: 'bg-blue-500', green: 'bg-green-500' }

interface StatusLightProps extends ComponentProps<'div'> {
  color: keyof typeof COLORS
  size?: keyof typeof sizes
}

const sizes = {
  sm: 'size-2',
  md: 'size-3',
  lg: 'size-4',
}

export function StatusLight({
  color,
  size = 'md',
  children,
  className,
  ...props
}: StatusLightProps) {
  return (
    <div className={cx('inline-flex items-center gap-2', className)} {...props} role="status">
      <span
        className={cx('inline-block shrink-0 grow-0 rounded-full', sizes[size], COLORS[color])}
        aria-label={color}
      />
      {children}
    </div>
  )
}
