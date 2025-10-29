import {
  FloatingPortal,
  Placement,
  autoUpdate,
  flip,
  offset,
  shift,
  useDismiss,
  useFloating,
  useFocus,
  useHover,
  useInteractions,
  useMergeRefs,
  useRole,
  useTransitionStyles,
} from '@floating-ui/react'
import {
  ComponentProps,
  HTMLProps,
  ReactNode,
  Ref,
  RefObject,
  cloneElement,
  createContext,
  isValidElement,
  useContext,
  useImperativeHandle,
  useMemo,
  useRef,
  useState,
} from 'react'

import { cx } from '@/utils/css.ts'

interface TooltipOptions {
  initialOpen?: boolean
  placement?: Placement
  open?: boolean
  onOpenChange?: (open: boolean) => void
}

export function useTooltipWithInteractions({
  initialOpen = false,
  placement = 'top',
  open: controlledOpen,
  onOpenChange: setControlledOpen,
}: TooltipOptions = {}) {
  const [uncontrolledOpen, setUncontrolledOpen] = useState(initialOpen)

  const open = controlledOpen ?? uncontrolledOpen
  const setOpen = setControlledOpen ?? setUncontrolledOpen

  const data = useFloating({
    placement,
    open,
    onOpenChange: setOpen,
    whileElementsMounted: autoUpdate,
    middleware: [
      offset(5),
      flip({
        crossAxis: placement.includes('-'),
        fallbackAxisSideDirection: 'start',
        padding: 5,
      }),
      shift({ padding: 5 }),
    ],
  })

  const context = data.context

  const hover = useHover(context, {
    move: false,
    enabled: controlledOpen == null,
  })
  const focus = useFocus(context, {
    enabled: controlledOpen == null,
  })
  const dismiss = useDismiss(context)
  const role = useRole(context, { role: 'tooltip' })

  const interactions = useInteractions([hover, focus, dismiss, role])

  return useMemo(
    () => ({
      open,
      setOpen,
      ...interactions,
      ...data,
    }),
    [open, setOpen, interactions, data],
  )
}

const TooltipContext = createContext<ReturnType<typeof useTooltipWithInteractions> | null>(null)

export const useTooltipContext = () => {
  const context = useContext(TooltipContext)

  if (context == null) {
    throw new Error('Tooltip components must be wrapped in <Tooltip />')
  }

  return context
}

export function Tooltip({
  children,
  refs,
  ...options
}: {
  children: ReactNode
  refs?: RefObject<ReturnType<typeof useTooltipWithInteractions>['refs'] | null>
} & TooltipOptions) {
  const tooltip = useTooltipWithInteractions(options)
  if (refs) {
    refs.current = tooltip.refs
  }
  return <TooltipContext.Provider value={tooltip}>{children}</TooltipContext.Provider>
}

interface TooltipTriggerProps extends HTMLProps<HTMLElement> {
  asChild?: boolean
  ref?: Ref<HTMLElement>
}

export function TooltipTrigger({
  children,
  asChild = false,
  ref: propRef,
  ...props
}: TooltipTriggerProps) {
  const context = useTooltipContext()
  const childrenRef = (children as { ref?: unknown }).ref
  const ref = useMergeRefs([context.refs.setReference, propRef, childrenRef as Ref<unknown>])

  // `asChild` allows the user to pass any element as the anchor
  if (asChild && isValidElement(children)) {
    return cloneElement(
      children,
      context.getReferenceProps({
        ref,
        ...props,
        ...(children.props as HTMLProps<Element>),
        // @ts-expect-error make ts happy
        'data-state': context.open ? 'open' : 'closed',
      }),
    )
  }

  return (
    <button
      ref={ref}
      // The user can style the trigger based on the state
      data-state={context.open ? 'open' : 'closed'}
      {...context.getReferenceProps(props)}
    >
      {children}
    </button>
  )
}

interface TooltipContentProps extends HTMLProps<HTMLDivElement> {
  portal?: boolean
  ref?: Ref<HTMLDivElement>
}

export function TooltipContent({
  className,
  style,
  portal = true,
  children,
  ref: propRef,
  ...props
}: TooltipContentProps) {
  const { context: floatingContext, ...context } = useTooltipContext()
  const ref = useMergeRefs([context.refs.setFloating, propRef])

  const { isMounted, styles: transitionStyles } = useTransitionStyles(floatingContext, {
    duration: {
      open: 150,
      close: 0,
    },
    initial: {
      opacity: 0,
    },
  })

  if (!isMounted) return null

  let content = (
    <div
      ref={ref}
      {...context.getFloatingProps(props)}
      data-side={context.placement}
      className={cx(
        'bg-popover text-popover-foreground pointer-events-none z-50 overflow-hidden rounded border px-3 py-1.5 text-sm shadow-md',
        className,
      )}
      style={{
        ...context.floatingStyles,
        ...transitionStyles,
        ...style,
        transitionProperty: 'transform, opacity',
      }}
    >
      {children}
    </div>
  )

  if (portal) {
    content = <FloatingPortal>{content}</FloatingPortal>
  }

  return content
}

export interface VirtualTooltipHandle {
  open: (target: HTMLElement, content: ReactNode) => void
  close: () => void
}

interface VirtualTooltipProps extends Omit<ComponentProps<typeof TooltipContent>, 'ref'> {
  delay?: number
  ref?: Ref<VirtualTooltipHandle>
}

export function VirtualTooltip({ delay = 300, ref: propRef, ...props }: VirtualTooltipProps) {
  const [open, setOpen] = useState(false)
  const [content, setContent] = useState<ReactNode>(null)
  const refs = useRef<ReturnType<typeof useTooltipWithInteractions>['refs']>(null)

  const tid = useRef<number | undefined>(undefined)

  useImperativeHandle(propRef, () => ({
    open: (target, content) => {
      if (tid.current) {
        clearTimeout(tid.current)
      }
      refs.current?.setReference(target)
      setContent(content)
      setOpen(true)
    },
    close: () => {
      // https://stackoverflow.com/questions/55550096/ts2322-type-timeout-is-not-assignable-to-type-number-when-running-unit-te
      tid.current = window.setTimeout(() => {
        refs.current?.setReference(null)
        setContent(null)
        setOpen(false)
      }, delay)
    },
  }))

  return (
    <Tooltip open={open} onOpenChange={setOpen} refs={refs}>
      <TooltipContent {...props}>{content}</TooltipContent>
    </Tooltip>
  )
}
