import {
  FloatingFocusManager,
  FloatingPortal,
  autoUpdate,
  flip,
  offset,
  size,
  useDismiss,
  useFloating,
  useId,
  useInteractions,
  useListNavigation,
  useRole,
  useTransitionStyles,
} from '@floating-ui/react'
import { Check as CheckIcon, ChevronDown } from 'lucide-react'
import { ChangeEvent, ComponentProps, Ref, useRef, useState } from 'react'

import { cx } from '@/utils/css.ts'

import { Button } from './button'
import { Input } from './input'

interface ComboBoxProps extends Omit<ComponentProps<'input'>, 'onChange' | 'prefix'> {
  options: string[]
  initialValue?: string
  value?: string
  onChange: (value?: string) => void
  placeholder?: string
}

export function ComboBox({
  options,
  initialValue,
  value: controlledValue,
  onChange,
  placeholder,
  ...props
}: ComboBoxProps) {
  const [open, setOpen] = useState(false)

  const [inputValue, setInputValue] = useState('')

  const [uncontrolledValue, setUncontrolledValue] = useState<string | undefined>(initialValue)
  const selectedValue = typeof initialValue === 'undefined' ? controlledValue : uncontrolledValue
  const setSelectedValue =
    typeof initialValue === 'undefined'
      ? onChange
      : (val?: string) => {
          onChange(val)
          setUncontrolledValue(val)
        }

  const [activeIndex, setActiveIndex] = useState<number | null>(null)
  const [filteredOptions, setFilteredOptions] = useState<string[]>([])
  const listRef = useRef<(HTMLElement | null)[]>([])

  const { refs, strategy, x, y, context } = useFloating<HTMLInputElement>({
    whileElementsMounted: autoUpdate,
    open,
    onOpenChange: setOpen,
    middleware: [
      offset(5),
      flip({ padding: 5 }),
      size({
        // NOTE: The availableHeight is calculated based on the visual viewport,
        // excluding the height occupied by the soft keyboard.
        apply({ rects, availableHeight, elements }) {
          Object.assign(elements.floating.style, {
            width: `${rects.reference.width.toString()}px`,
            maxHeight: `${(availableHeight * 0.75).toString()}px`,
          })
        },
        padding: 5,
      }),
    ],
  })

  const { isMounted, styles: transitionStyles } = useTransitionStyles(context, {
    duration: {
      open: 300,
      close: 0,
    },
    initial: ({ side }) => ({
      opacity: 0,
      transform: transforms[side],
    }),
  })

  const role = useRole(context, { role: 'listbox' })
  const dismiss = useDismiss(context, {
    outsidePress: (event) => {
      return !(event.target as HTMLElement).classList.contains('suffix')
    },
  })
  const listNav = useListNavigation(context, {
    listRef,
    activeIndex,
    onNavigate: setActiveIndex,
    virtual: true,
    loop: true,
  })

  const { getReferenceProps, getFloatingProps, getItemProps } = useInteractions([
    role,
    dismiss,
    listNav,
  ])

  return (
    <>
      <Input
        className="placeholder-muted w-[250px]"
        spellCheck={false}
        placeholder={placeholder}
        aria-autocomplete="list"
        {...getReferenceProps({
          ref: refs.setReference,
          value: inputValue,
          onChange: (event: ChangeEvent<HTMLInputElement>) => {
            const value = event.target.value
            setOpen(true)
            setInputValue(value)
            setActiveIndex(0)
            setFilteredOptions(
              options.filter((option) => option.toLowerCase().includes(value.toLowerCase())),
            )

            if (!value) {
              setSelectedValue(undefined)
            }
          },
          onKeyDown: (event) => {
            if (event.key === 'Enter' && activeIndex != null && filteredOptions[activeIndex]) {
              setOpen(false)
              setInputValue(filteredOptions[activeIndex])
              setSelectedValue(filteredOptions[activeIndex])
              setActiveIndex(null)
            }
          },
          onBlur: () => {
            setInputValue(selectedValue ?? '')
          },
        })}
        suffix={
          <span className="right-0! inline-block h-10">
            <Button
              variant="ghost"
              size="sm"
              className="suffix group-has-[:focus-visible]:border-l-ring pointer-events-auto h-full rounded-none rounded-r-md border-l px-2! focus-visible:border-l-transparent!"
              onClick={(event) => {
                event.preventDefault()
                if (!open) {
                  setFilteredOptions(options)
                  setOpen(true)
                  setActiveIndex(0)
                } else {
                  setOpen(false)
                  setActiveIndex(null)
                }
                // NOTE: The `focus` method must be placed within a `setTimeout`;
                // otherwise, the `open` state above will always remain `false`.
                setTimeout(() => {
                  refs.domReference.current?.focus()
                })
              }}
            >
              <ChevronDown className="pointer-events-none size-5" />
            </Button>
          </span>
        }
        {...props}
      />
      {isMounted && filteredOptions.length !== 0 && (
        <FloatingPortal>
          <FloatingFocusManager context={context} initialFocus={-1} visuallyHiddenDismiss>
            <ul
              className="scrollbar-none bg-popover z-50 space-y-1 overflow-y-auto border shadow"
              {...getFloatingProps({
                ref: refs.setFloating,
                style: {
                  position: strategy,
                  left: x,
                  top: y,
                  ...transitionStyles,
                },
              })}
            >
              {filteredOptions.map((item, index) => (
                <Item
                  key={item}
                  {...getItemProps({
                    ref(node) {
                      listRef.current[index] = node
                    },
                    onClick() {
                      setOpen(false)
                      setInputValue(item)
                      setSelectedValue(item)
                      setActiveIndex(null)
                      refs.domReference.current?.focus()
                    },
                  })}
                  active={activeIndex === index}
                >
                  <span className="truncate">{item}</span>
                  {selectedValue === item && (
                    <span className="text-primary ml-1">
                      <CheckIcon className="size-4" />
                    </span>
                  )}
                </Item>
              ))}
            </ul>
          </FloatingFocusManager>
        </FloatingPortal>
      )}
    </>
  )
}

interface ItemProps extends ComponentProps<'li'> {
  active?: boolean
  ref?: Ref<HTMLLIElement>
}

function Item({ children, active = false, ref, ...props }: ItemProps) {
  const id = useId()
  return (
    <li
      ref={ref}
      id={id}
      className={cx('relative', { 'bg-accent': active })}
      role="option"
      aria-selected={active}
      {...props}
    >
      <Button variant="ghost" className="w-full max-w-full justify-between!">
        {children}
      </Button>
    </li>
  )
}

const transforms = {
  left: 'translateX(10px)',
  right: 'translateX(-10px)',
  top: 'translateY(10px)',
  bottom: 'translateY(-10px)',
}
