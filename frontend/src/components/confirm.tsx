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

import { Button } from './button'
import { Dialog, DialogClose, DialogContent, DialogDescription, DialogHeading } from './dialog'

interface ConfirmOption {
  heading: ReactNode
  description?: ReactNode
  okText?: string
  cancelText?: string
  onOk?: () => void | Promise<void>
  onCancel?: () => void | Promise<void>
  cancelButtonVariant?: ComponentProps<typeof Button>['variant']
  oKButtonVariant?: ComponentProps<typeof Button>['variant']
  cancelButtonClassName?: string
  oKButtonClassName?: string
}

interface ConfirmHandle {
  open: (option: ConfirmOption) => void
}

interface ConfirmProps extends Omit<ComponentProps<typeof DialogContent>, 'ref'> {
  ref?: Ref<ConfirmHandle>
}

export function Confirm({ ref, ...props }: ConfirmProps) {
  const [open, setOpen] = useState(false)

  const [options, setOptions] = useState({} as ConfirmOption)
  const {
    heading,
    description,
    onOk,
    onCancel,
    okText = 'OK',
    cancelText = 'Cancel',
    oKButtonVariant = 'destructive',
    cancelButtonVariant = 'outline',
    oKButtonClassName,
    cancelButtonClassName,
  } = options

  const [pending, setPending] = useState(false)

  useImperativeHandle(ref, () => ({
    open: (options) => {
      setOptions(options)
      setOpen(true)
    },
  }))

  const handleOk = async () => {
    setPending(true)
    try {
      if (onOk) {
        await onOk()
      }
    } catch (err) {
      console.error(err)
    } finally {
      setPending(false)
      setOpen(false)
    }
  }

  const handleCancel = async () => {
    if (onCancel) {
      await onCancel()
    }
    setOpen(false)
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent {...props}>
        <DialogHeading>{heading}</DialogHeading>
        {description && <DialogDescription>{description}</DialogDescription>}
        <DialogClose />
        <div className="flex justify-end space-x-3">
          <Button
            variant={cancelButtonVariant}
            className={cancelButtonClassName}
            onClick={() => void handleCancel()}
          >
            {cancelText}
          </Button>
          <Button
            variant={oKButtonVariant}
            className={oKButtonClassName}
            disabled={pending}
            onClick={() => void handleOk()}
          >
            {okText}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}

const ConfirmContext = createContext<RefObject<ConfirmHandle> | null>(null)

export function ConfirmProvider({ children }: { children: ReactNode }) {
  const ref = useRef<ConfirmHandle>(null!)

  return (
    <ConfirmContext value={ref}>
      <>
        <Confirm ref={ref} />
        {children}
      </>
    </ConfirmContext>
  )
}

export function useConfirm() {
  const confirmRef = useContext(ConfirmContext)

  return useMemo(
    () => ({
      open: (option: ConfirmOption) => {
        confirmRef?.current.open(option)
      },
    }),
    [confirmRef],
  )
}
