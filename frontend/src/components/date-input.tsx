import { ComponentProps, SyntheticEvent, useEffect, useRef, useState } from 'react'

import { cx } from '../utils/css.ts'
import { Input } from './input'

interface DateInputProps extends Omit<ComponentProps<'div'>, 'onChange'> {
  initialValue?: string
  onChange?: (value: string) => void
  placeholders?: [string, string, string]
}

export function DateInput({
  initialValue = '',
  onChange,
  placeholders = ['YYYY', 'MM', 'DD'],
  className,
}: DateInputProps) {
  const [[initialYear, initialMonth, initialDay], _] = useState(
    () => parseDateStr(initialValue) ?? [],
  )
  const [year, setYear] = useState(initialYear ?? '')
  const [month, setMonth] = useState(initialMonth ?? '')
  const [day, setDay] = useState(initialDay ?? '')

  const yearRef = useRef<HTMLInputElement>(null)
  const monthRef = useRef<HTMLInputElement>(null)
  const dayRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (isValidDateStr(year, month, day)) {
      onChange?.(normalizeDateStr(year, month, day))
    }
  }, [year, month, day, onChange])

  const handleSelect = (event: SyntheticEvent<HTMLInputElement>) => {
    // https://stackoverflow.com/questions/511088/use-javascript-to-place-cursor-at-end-of-text-in-text-input-element
    // https://stackoverflow.com/questions/66710037/property-selectionstart-does-not-exist-on-type-eventtarget
    const el = event.target as HTMLInputElement
    setTimeout(function () {
      el.selectionStart = el.selectionEnd = 10000
    }, 0)
  }

  const fixLastDayOfMonth = (year: string, month: string, day: string) => {
    if (day) {
      const lastDay = getDaysInMonth(parseInt(year), parseInt(month))
      if (parseInt(day) > lastDay) {
        setDay(lastDay.toString())
      }
    }
  }

  const inputStyles =
    'select-none tabular-nums placeholder-muted caret-[transparent] selection:bg-[0_0]'

  return (
    <div className={cx('inline-flex items-center', className)}>
      <Input
        ref={yearRef}
        className={cx(inputStyles, 'w-[4rem]!')}
        type="text"
        inputMode="numeric"
        placeholder={placeholders[0]}
        value={year}
        onChange={(event) => {
          let value = event.target.value.trim()

          if (value === '' || value === '0') {
            setYear('')
            return
          }

          if (!Number.isInteger(Number(value))) return

          if (value === '0000') value = '000'

          if (value.length > 4) value = value[value.length - 1]

          setYear(value)
          fixLastDayOfMonth(value, month, day)

          if (value.length === 4) {
            monthRef.current?.focus()
          }
        }}
        onPaste={(event) => {
          event.preventDefault()
          const value = event.clipboardData.getData('text').trim()
          const rv = parseDateStr(value)
          if (rv) {
            setYear(rv[0])
            setMonth(rv[1])
            setDay(rv[2])
          }
        }}
        onSelect={handleSelect}
      />
      <span className="divider" />
      <Divider />
      <Input
        ref={monthRef}
        className={cx(inputStyles, 'w-[3.1rem]!')}
        type="text"
        inputMode="numeric"
        placeholder={placeholders[1]}
        value={month}
        onChange={(event) => {
          let value = event.target.value.trim()

          if (value.length === 0) {
            setMonth('')
            return
          }

          if (!Number.isInteger(Number(value))) return

          if (value === '00') value = '0'

          if (value.length > 2) value = value[value.length - 1]

          if (value.length === 2 && parseInt(value) > 12) value = value[value.length - 1]

          setMonth(value)
          fixLastDayOfMonth(year, value, day)

          if (value !== '0' && value !== '1') {
            dayRef.current?.focus()
          }
        }}
        onPaste={(event) => {
          event.preventDefault()
        }}
        onSelect={handleSelect}
      />
      <Divider />
      <Input
        ref={dayRef}
        className={cx(inputStyles, 'w-[3.1rem]!')}
        type="text"
        inputMode="numeric"
        placeholder={placeholders[2]}
        value={day}
        onChange={(event) => {
          let value = event.target.value.trim()

          if (value.length === 0) {
            setDay('')
            return
          }

          if (!Number.isInteger(Number(value))) return

          if (value === '00') value = '0'

          if (value.length > 2) value = value[value.length - 1]

          if (value.length === 2) {
            const days = getDaysInMonth(parseInt(year), parseInt(month))
            if (parseInt(value) > days) {
              value = value[value.length - 1]
            }
          }
          setDay(value)
        }}
        onPaste={(event) => {
          // https://stackoverflow.com/questions/686995/catch-paste-input
          event.preventDefault()
        }}
        onSelect={handleSelect}
      />
    </div>
  )
}

function Divider() {
  return (
    <span className="bg-muted-foreground/70 mx-0.5 inline-block h-[1px] w-3 -rotate-[60deg] whitespace-pre select-none" />
  )
}

const DAYS_OF_MONTH = {
  2: 28,

  1: 31,
  3: 31,
  5: 31,
  7: 31,
  8: 31,
  10: 31,
  12: 31,

  4: 30,
  6: 30,
  9: 30,
  11: 30,
}

type MonthType = keyof typeof DAYS_OF_MONTH

function getDaysInMonth(year: number, month: number) {
  if (year && month) return new Date(year, month, 0).getDate()
  if (month) return DAYS_OF_MONTH[month as MonthType]

  return 31
}

const dateReg = /^(\d{4})-(\d{1,2})-(\d{1,2})$/

function parseDateStr(str: string): [string, string, string] | null {
  if (!str) return null

  const match = dateReg.exec(str)
  return match ? [match[1], match[2], match[3]] : null
}

function isValidDateStr(...args: string[]): boolean {
  if (args.length === 1) return dateReg.test(args[0])
  if (args.length === 3) return dateReg.test(args.join('-'))

  return false
}

function normalizeDateStr(year: string, month: string, day: string): string {
  if (parseInt(year) === 0) year = '1'
  if (parseInt(month) === 0) month = '1'
  if (parseInt(day) === 0) day = '1'

  return `${year.padStart(4, '0')}-${month.padStart(2, '0')}-${day.padStart(2, '0')}`
}
