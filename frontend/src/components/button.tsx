import React, { HTMLProps, Ref } from 'react'

import { cx } from '../utils/css.ts'

const baseStyles =
  'inline-flex items-center justify-center text-sm font-medium ' +
  'whitespace-nowrap rounded transition-colors animate-pulsate ' +
  'ring-offset-background focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-offset-0 ' +
  'disabled:pointer-events-none disabled:opacity-50'

const variants = {
  primary: 'bg-primary text-primary-foreground hover:bg-primary/90',
  secondary: 'bg-secondary text-secondary-foreground hover:bg-secondary/80',
  destructive: 'bg-destructive text-destructive-foreground hover:bg-destructive/90',
  outline: 'border border-input bg-background hover:bg-accent hover:text-accent-foreground',
  ghost: 'hover:bg-accent hover:text-accent-foreground',
  link: 'text-primary underline-offset-4 hover:underline',
}

const sizes = {
  sm: 'h-9 px-3',
  md: 'h-10 px-4 py-2',
  lg: 'h-11 px-8',
  icon: 'size-10',
}

interface ButtonProps extends Omit<HTMLProps<HTMLElement>, 'size'> {
  variant?: keyof typeof variants
  size?: keyof typeof sizes
  loading?: boolean
  tag?: string
  ref?: Ref<HTMLElement>
}

export function Button({
  className,
  variant = 'primary',
  size = 'md',
  tag = 'button',
  children,
  ref,
  ...rest
}: ButtonProps) {
  const props = {
    ref,
    className: cx(baseStyles, variants[variant], sizes[size], className),
    // https://dev.to/tylerjdev/when-role-button-is-not-enough-dac
    // NOTE: It is not possible to fully make other elements like `span` have the default behavior of `button`,
    // such as when pressing space or enter.
    role: 'button',
    ...rest,
  }
  return React.createElement(tag, props, children)
}
