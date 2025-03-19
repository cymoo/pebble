import {
  ComponentProps,
  KeyboardEvent,
  ReactElement,
  ReactNode,
  cloneElement,
  useId,
  useState,
} from 'react'

import { cx } from '@/utils/css.ts'

import { Button } from './button'

interface RadioButtonProps<T> extends Omit<ComponentProps<'div'>, 'onChange'> {
  name?: string
  initialValue?: T | null
  value?: T | null
  onChange: (value: T | null) => void
  options: { label: ReactNode; value: T }[]
  selfToggleable?: boolean
  renderLabel?: (option: { label: ReactNode; value: T }) => ReactElement
}

export function RadioButton<T extends ComponentProps<'input'>['value']>({
  name,
  initialValue,
  value: controlledValue,
  onChange,
  options,
  selfToggleable = false,
  renderLabel,
  className,
  ...props
}: RadioButtonProps<T>) {
  const id = useId()

  const [uncontrolledValue, setUncontrolledValue] = useState<T | null | undefined>(initialValue)
  const value = typeof initialValue === 'undefined' ? controlledValue : uncontrolledValue
  const setValue =
    typeof initialValue === 'undefined'
      ? onChange
      : (val: T | null) => {
          onChange(val)
          setUncontrolledValue(val)
        }

  const getLabelProps = (option: { label: ReactNode; value: T }) => ({
    htmlFor: `${id}-${String(option.value)}`,
    tabIndex: 0,
    // NOTE: It is advisable to avoid using `onKeyUp`, as it may trigger when focusing on `PopoverTrigger` and pressing `Enter`,
    // especially if the component is the first focusable element.
    onKeyDown: (event: KeyboardEvent) => {
      console.log('event: ', event)
      if (event.key === 'Enter' || event.key === ' ') {
        if (selfToggleable && option.value === value) {
          setValue(null)
        } else {
          setValue(option.value)
        }
      }
    },
  })

  return (
    <div className={cx('flex gap-5', className)} {...props}>
      {options.map((option) => (
        <div className="inline-block" key={String(option.value)}>
          <input
            type="radio"
            id={`${id}-${String(option.value)}`}
            className="sr-only"
            name={name ?? id}
            tabIndex={-1}
            value={option.value}
            checked={value === option.value}
            onClick={() => {
              if (selfToggleable && option.value === value) {
                setValue(null)
              }
            }}
            onChange={() => {
              setValue(option.value)
            }}
          />
          {renderLabel ? (
            cloneElement(renderLabel(option), getLabelProps(option))
          ) : (
            <Button
              tag="label"
              variant={value === option.value ? 'primary' : 'outline'}
              className="border-input border"
              {...getLabelProps(option)}
            >
              {option.label}
            </Button>
          )}
        </div>
      ))}
    </div>
  )
}
