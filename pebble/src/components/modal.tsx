import {
  ComponentProps,
  ReactNode,
  Ref,
  RefObject,
  createContext,
  useContext,
  useImperativeHandle,
  useMemo,
  useRef,
  useState,
} from 'react'

import { IS_TOUCH_DEVICE } from '@/utils/browser.ts'
import { cx } from '@/utils/css.ts'

import { Dialog, DialogClose, DialogContent, DialogHeading } from './dialog'

interface ModalOption {
  heading?: ReactNode
  headingVisible?: boolean
  content: ReactNode
}

interface ModalHandle {
  open: (option: ModalOption) => void
  close: () => void
}

interface ModalProps extends Omit<ComponentProps<typeof DialogContent>, 'ref'> {
  ref?: Ref<ModalHandle>
}

export function Modal({ ref, ...props }: ModalProps) {
  const [open, setOpen] = useState(false)

  const [options, setOptions] = useState({} as ModalOption)
  const { heading, content, headingVisible = false } = options

  useImperativeHandle(ref, () => ({
    open: (options) => {
      setOptions(options)
      setOpen(true)
    },
    close: () => {
      setOpen(false)
    },
  }))

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent {...props}>
        {heading && (
          <DialogHeading className={cx({ 'sr-only': !headingVisible })}>{heading}</DialogHeading>
        )}
        <DialogClose className="sr-only" />
        {content}
      </DialogContent>
    </Dialog>
  )
}

const ModalContext = createContext<RefObject<ModalHandle> | null>(null)

export function ModalProvider({ children }: { children: ReactNode }) {
  const ref = useRef<ModalHandle>(null!)

  return (
    <ModalContext value={ref}>
      <>
        <Modal
          ref={ref}
          // animation={false}
          overlayClassName={cx('bg-black/90', { 'items-end!': IS_TOUCH_DEVICE })}
          className={cx('max-w-[640px]! p-4!', IS_TOUCH_DEVICE ? 'max-h-[80vh]' : 'max-h-[640px]')}
        />
        {children}
      </>
    </ModalContext>
  )
}

export function useModal() {
  const modalRef = useContext(ModalContext)

  return useMemo(
    () => ({
      open: (option: ModalOption) => {
        modalRef?.current.open(option)
      },
      close: () => {
        modalRef?.current.close()
      },
    }),
    [modalRef],
  )
}
