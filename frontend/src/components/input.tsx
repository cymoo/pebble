import { ComponentProps, HTMLAttributes, ReactElement, Ref, cloneElement } from 'react'

import { cx } from '@/utils/css.ts'

export type InputProps = Omit<ComponentProps<'input'>, 'prefix'> & {
  prefix?: ReactElement<HTMLAttributes<HTMLElement>>
  suffix?: ReactElement<HTMLAttributes<HTMLElement>>
  ref?: Ref<HTMLInputElement>
}

const inputStyles =
  'flex h-10 w-full rounded border border-input bg-background transition-shadow ' +
  'px-3 py-2 text-sm ring-offset-background ' +
  'file:border-0 file:bg-transparent file:text-sm file:font-medium ' +
  'placeholder:text-muted-foreground ' +
  'focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-offset-0 ' +
  'disabled:cursor-not-allowed disabled:opacity-50'

const iconStyles =
  'absolute top-1/2 -translate-y-1/2 text-muted-foreground ' +
  'pointer-events-none group-focus-within:text-ring leading-none'

export function Input({ className, prefix, suffix, ref, ...props }: InputProps) {
  if (!prefix && !suffix) {
    return <input ref={ref} className={cx(inputStyles, className)} {...props} />
  }

  return (
    <span className="group relative inline-block">
      {prefix &&
        cloneElement(prefix, {
          className: cx(iconStyles, 'left-3', prefix.props.className),
          'aria-hidden': 'true',
        })}
      <input
        className={cx(inputStyles, { 'pl-10': !!prefix, 'pr-10': !!suffix }, className)}
        ref={ref}
        {...props}
      />
      {suffix &&
        cloneElement(suffix, {
          className: cx(iconStyles, 'right-3', suffix.props.className),
          'aria-hidden': 'true',
        })}
    </span>
  )
}
