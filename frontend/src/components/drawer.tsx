import {
  FloatingFocusManager,
  FloatingOverlay,
  FloatingPortal,
  useMergeRefs,
} from '@floating-ui/react'
import { animated, useSpring } from '@react-spring/web'
import { ComponentProps, Ref, useLayoutEffect, useRef, useState } from 'react'

import { IS_IOS } from '@/utils/browser.ts'
import { cx } from '@/utils/css.ts'
import { useIsUnmounted } from '@/utils/hooks/use-unmount.ts'

import {
  Dialog,
  DialogClose,
  DialogDescription,
  DialogHeading,
  DialogTrigger,
  useDialogContext,
} from './dialog'

export const Drawer = Dialog
export const DrawerTrigger = DialogTrigger
export const DrawerClose = DialogClose
export const DrawerHeading = DialogHeading
export const DrawerDescription = DialogDescription

const sides = {
  top: 'inset-x-0 top-0 border-b',
  bottom: 'inset-x-0 bottom-0 border-t',
  left: 'inset-y-0 left-0 h-full w-3/4 border-r max-w-xl',
  right: 'inset-y-0 right-0 h-full w-3/4 border-l max-w-xl',
}

interface DrawerContentProps extends ComponentProps<'div'> {
  side?: keyof typeof sides
  overlayClassName?: string
  animation?: boolean
  afterEnter?: () => void
  afterLeave?: () => void
  alwaysRender?: boolean
  destroyOnClose?: boolean
  ref?: Ref<HTMLDivElement>
}

export function DrawerContent({
  side = 'right',
  overlayClassName,
  animation = true,
  alwaysRender = false,
  destroyOnClose = false,
  afterEnter,
  afterLeave,
  className,
  style = {},
  ref: propRef,
  ...props
}: DrawerContentProps) {
  const { context: floatingContext, open, ...context } = useDialogContext()
  const ref = useMergeRefs([context.refs.setFloating, propRef])

  const [active, setActive] = useState(open)

  useLayoutEffect(() => {
    if (open) {
      setActive(true)
    }
  }, [open])

  const isUnmounted = useIsUnmounted()

  const { opacity, percent } = useSpring({
    opacity: open ? 1 : 0,
    percent: open ? 0 : 100,
    immediate: !animation,
    onStart: () => {
      setActive(true)
    },
    onRest: () => {
      if (isUnmounted()) return
      setActive(open)

      if (open) {
        afterEnter?.()
      } else {
        afterLeave?.()
      }
    },
  })

  const shouldRender = useShouldRender(active, alwaysRender, destroyOnClose)

  if (!shouldRender) return null

  const AnimatedFloatingOverlay = animated(FloatingOverlay)
  const AnimatedDiv = animated('div')

  return (
    <FloatingPortal>
      <AnimatedFloatingOverlay
        className={cx('z-50 flex items-center justify-center bg-black/80', overlayClassName)}
        // NOTE: On iOS, when the drawer is closed, the body exhibits unusual scrolling behavior.
        lockScroll={!IS_IOS ? active : open}
        style={{ opacity, display: active ? 'flex' : 'none' }}
      >
        <FloatingFocusManager context={floatingContext}>
          <AnimatedDiv
            ref={ref}
            {...context.getFloatingProps(props)}
            aria-labelledby={context.labelId}
            aria-describedby={context.descriptionId}
            className={cx(
              'bg-background fixed p-6 shadow-lg ease-out rounded',
              sides[side],
              className,
            )}
            style={{
              ...style,
              transform: percent.to((v) => {
                if (!v) return 'none'

                switch (side) {
                  case 'bottom':
                    return `translate(0, ${v.toString()}%)`
                  case 'top':
                    return `translate(0, -${v.toString()}%)`
                  case 'left':
                    return `translate(-${v.toString()}%, 0)`
                  case 'right':
                  default:
                    return `translate(${v.toString()}%, 0)`
                }
              }),
            }}
          >
            {props.children}
          </AnimatedDiv>
        </FloatingFocusManager>
      </AnimatedFloatingOverlay>
    </FloatingPortal>
  )
}

const useShouldRender = (active: boolean, alwaysRender: boolean, destroyOnClose: boolean) => {
  const initialized = useInitialized(active)
  if (alwaysRender) return true
  if (active) return true
  if (!initialized) return false
  return !destroyOnClose
}

const useInitialized = (check: boolean) => {
  const initializedRef = useRef(check)
  if (check) {
    initializedRef.current = true
  }
  return initializedRef.current
}
