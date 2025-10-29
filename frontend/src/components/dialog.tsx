import {
  FloatingFocusManager,
  FloatingOverlay,
  FloatingPortal,
  useClick,
  useDismiss,
  useFloating,
  useId,
  useInteractions,
  useMergeRefs,
  useRole,
  useTransitionStyles,
} from '@floating-ui/react'
import { X } from 'lucide-react'
import {
  ComponentProps,
  HTMLProps,
  ReactNode,
  Ref,
  cloneElement,
  createContext,
  isValidElement,
  useContext,
  useLayoutEffect,
  useMemo,
  useState,
} from 'react'

import { cx } from '@/utils/css.ts'
import { omit } from '@/utils/obj'

import { Button } from './button'

interface DialogOptions {
  initialOpen?: boolean
  open?: boolean
  onOpenChange?: (open: boolean) => void
  // when press escape-key or click mask
  beforeDismiss?: (close: () => void) => void
  outsidePress?: boolean | ((event: MouseEvent) => boolean)
}

export function useDialogWithInteractions({
  initialOpen = false,
  open: controlledOpen,
  onOpenChange: setControlledOpen,
  beforeDismiss,
  outsidePress,
}: DialogOptions = {}) {
  const [uncontrolledOpen, setUncontrolledOpen] = useState(initialOpen)
  const [labelId, setLabelId] = useState<string | undefined>()
  const [descriptionId, setDescriptionId] = useState<string | undefined>()

  const open = controlledOpen ?? uncontrolledOpen
  const setOpen = setControlledOpen ?? setUncontrolledOpen

  const data = useFloating({
    open,
    onOpenChange(nextOpen, event, reason) {
      if (reason === 'escape-key') {
        if (window.pswp) {
          window.pswp.close()
          return
        }
      }

      if (
        typeof beforeDismiss === 'function' &&
        // Other ones include 'reference-press' and 'ancestor-scroll' if enabled.
        (reason === 'escape-key' || reason === 'outside-press')
      ) {
        beforeDismiss(() => {
          setOpen(false)
        })
        return
      }

      setOpen(nextOpen)
    },
  })

  const context = data.context

  const click = useClick(context, {
    enabled: controlledOpen == null,
  })
  const dismiss = useDismiss(context, {
    outsidePressEvent: 'mousedown',
    outsidePress,
  })
  const role = useRole(context)

  const interactions = useInteractions([click, dismiss, role])

  return useMemo(
    () => ({
      open,
      setOpen,
      ...interactions,
      ...data,
      labelId,
      descriptionId,
      setLabelId,
      setDescriptionId,
    }),
    [open, setOpen, interactions, data, labelId, descriptionId],
  )
}

const DialogContext = createContext<ReturnType<typeof useDialogWithInteractions> | null>(null)

export function useDialogContext() {
  const context = useContext(DialogContext)

  if (context == null) {
    throw new Error('Dialog components must be wrapped in <Dialog />')
  }

  return context
}

export function Dialog({
  children,
  ...options
}: {
  children: ReactNode
} & DialogOptions) {
  const dialog = useDialogWithInteractions(options)
  return <DialogContext.Provider value={dialog}>{children}</DialogContext.Provider>
}

interface DialogTriggerProps extends HTMLProps<HTMLElement> {
  children: ReactNode
  asChild?: boolean
  ref?: Ref<HTMLElement>
}

export function DialogTrigger({
  children,
  asChild = false,
  ref: propRef,
  ...props
}: DialogTriggerProps) {
  const context = useDialogContext()
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
    <Button
      ref={ref}
      // The user can style the trigger based on the state
      data-state={context.open ? 'open' : 'closed'}
      {...context.getReferenceProps(props)}
    >
      {children}
    </Button>
  )
}

interface DialogContentProps extends ComponentProps<'div'> {
  overlayClassName?: string
  animation?: boolean
  ref?: Ref<HTMLDivElement>
}

export function DialogContent({
  overlayClassName,
  animation = true,
  className,
  style,
  children,
  ref: propRef,
  ...props
}: DialogContentProps) {
  const { context: floatingContext, ...context } = useDialogContext()
  const ref = useMergeRefs([context.refs.setFloating, propRef])

  const { isMounted, styles: transitionStyles } = useTransitionStyles(floatingContext, {
    duration: 300,
    initial: {
      opacity: 0,
      transform: 'scale(0.93)',
    },
  })

  if (!isMounted) return null

  return (
    <FloatingPortal>
      <FloatingOverlay
        className={cx('z-50 flex items-center justify-center bg-black/80', overlayClassName)}
        lockScroll
        style={omit(transitionStyles, 'transform')}
      >
        <FloatingFocusManager context={floatingContext}>
          <div
            ref={ref}
            className={cx(
              'bg-background relative flex w-full max-w-lg flex-col gap-4 border p-6 shadow-lg rounded',
              className,
            )}
            style={Object.assign(animation ? omit(transitionStyles, 'opacity') : {}, style)}
            aria-labelledby={context.labelId}
            aria-describedby={context.descriptionId}
            {...context.getFloatingProps(props)}
          >
            {children}
          </div>
        </FloatingFocusManager>
      </FloatingOverlay>
    </FloatingPortal>
  )
}

export function DialogHeading({ children, className, ref, ...props }: ComponentProps<'h2'>) {
  const { setLabelId } = useDialogContext()
  const id = useId()

  // Only sets `aria-labelledby` on the Dialog root element
  // if this component is mounted inside it.
  useLayoutEffect(() => {
    setLabelId(id)
    return () => {
      setLabelId(undefined)
    }
  }, [id, setLabelId])

  return (
    <h2
      ref={ref}
      id={id}
      className={cx('text-lg leading-none font-semibold tracking-tight', className)}
      {...props}
    >
      {children}
    </h2>
  )
}

export function DialogDescription({ children, className, ref, ...props }: ComponentProps<'p'>) {
  const { setDescriptionId } = useDialogContext()
  const id = useId()

  // Only sets `aria-describedby` on the Dialog root element
  // if this component is mounted inside it.
  useLayoutEffect(() => {
    setDescriptionId(id)
    return () => {
      setDescriptionId(undefined)
    }
  }, [id, setDescriptionId])

  return (
    <div ref={ref} id={id} className={cx('text-muted-foreground text-sm', className)} {...props}>
      {children}
    </div>
  )
}

export function DialogClose({ className, ref, ...props }: ComponentProps<typeof Button>) {
  const { setOpen } = useDialogContext()
  return (
    <Button
      ref={ref}
      type="button"
      variant="ghost"
      className={cx('absolute top-4 right-4', className)}
      onClick={() => {
        setOpen(false)
      }}
      {...props}
    >
      <X className="size-4" />
      <span className="sr-only">Close</span>
    </Button>
  )
}
