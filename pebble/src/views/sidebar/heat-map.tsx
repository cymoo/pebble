import React, { ComponentProps, useRef } from 'react'
import useSWR from 'swr'

import { cx } from '@/utils/css.ts'
import { formatDate, getDatesBetween, getNextSunday, getPreviousMonday } from '@/utils/date.ts'

import { VirtualTooltip, VirtualTooltipHandle } from '@/components/tooltip.tsx'

import { GET_DAILY_POST_COUNTS } from '@/api.ts'

import { useTheme } from './theme-toggle.tsx'

interface HeatMapProps extends ComponentProps<'div'> {
  startDate: Date
  endDate?: Date
}

export const HeatMap = React.memo(function HeatMap({
  startDate,
  endDate = new Date(),
  className,
  ...props
}: HeatMapProps) {
  const { theme } = useTheme()

  const firstMonday = getPreviousMonday(startDate)
  const lastSunday = getNextSunday(endDate)
  const dates = getDatesBetween(firstMonday, lastSunday).map((date) => formatDate(date))

  const startDateStr = formatDate(firstMonday)
  const endDateStr = formatDate(lastSunday)
  const offset = (-new Date().getTimezoneOffset()).toString()
  const { data: counts } = useSWR<number[]>(
    `${GET_DAILY_POST_COUNTS}?start_date=${startDateStr}&end_date=${endDateStr}&offset=${offset}`,
  )

  const today = formatDate()

  const ring = 'ring-ring ring-offset-1 ring-1'

  return (
    <div
      className={cx('grid grid-flow-col grid-rows-7 gap-1 *:aspect-square', className)}
      role="grid"
      aria-label="heatmap of activities"
      {...props}
    >
      {dates.map((date, idx) => {
        const count = counts?.[idx] ?? 0
        return (
          <a
            key={date}
            className={cx('bg-border cursor-pointer rounded', getColor(count, theme), {
              [ring]: date === today,
            })}
            data-date={date}
            data-count={count}
            role="gridcell"
            aria-label={`date: ${date}, activities: ${String(count)}`}
          />
        )
      })}
    </div>
  )
})

export function HeatMapWithTooltip({ className, ...props }: ComponentProps<typeof HeatMap>) {
  const ref = useRef<VirtualTooltipHandle>(null)

  return (
    <>
      <VirtualTooltip ref={ref} />
      <HeatMap
        className={cx('cursor-pointer', className)}
        onMouseOver={(event) => {
          if (event.target instanceof HTMLElement && event.target.tagName === 'A') {
            const { date, count } = event.target.dataset
            if (!date || !count) return
            const content = `${date}: ${count} memos`
            ref.current?.open(event.target, content)
          }
        }}
        onMouseOut={(event) => {
          if (event.target instanceof HTMLElement && event.target.tagName === 'A') {
            ref.current?.close()
          }
        }}
        {...props}
      />
    </>
  )
}

const getColor = (count: number, theme = 'light'): string => {
  if (theme === 'dark') {
    if (count >= 9) {
      return 'bg-green-300'
    } else if (count >= 7) {
      return 'bg-green-400'
    } else if (count >= 5) {
      return 'bg-green-500'
    } else if (count >= 3) {
      return 'bg-green-600'
    } else if (count >= 1) {
      return 'bg-green-700'
    } else {
      return 'bg-border'
    }
  } else {
    if (count >= 9) {
      return 'bg-green-700'
    } else if (count >= 7) {
      return 'bg-green-600'
    } else if (count >= 5) {
      return 'bg-green-500'
    } else if (count >= 3) {
      return 'bg-green-400'
    } else if (count >= 1) {
      return 'bg-green-300'
    } else {
      return 'bg-border'
    }
  }
}
