import { ComponentProps } from 'react'
import useSWR from 'swr'

import { cx } from '@/utils/css.ts'

import { GET_OVERALL_COUNTS } from '@/api.ts'

export function Stats({ className, ...props }: ComponentProps<'section'>) {
  const { data } = useSWR<{ post_count: number; tag_count: number; day_count: number }>(
    GET_OVERALL_COUNTS,
  )

  return (
    <section
      className={cx('flex items-center justify-between', className)}
      aria-label="statistics"
      {...props}
    >
      <StatItem label="MEMO" count={data?.post_count || '-'} />
      <StatItem label="TAG" count={data?.tag_count || '-'} />
      <StatItem label="DAY" count={data?.day_count || '-'} />
    </section>
  )
}

function StatItem({ label, count }: { label: string; count: number | '-' }) {
  return (
    <div className="flex flex-col items-center">
      <span className="text-lg font-semibold" aria-label={`${label} count: ${String(count)}`}>
        {count}
      </span>
      <span className="text-muted-foreground/75 text-sm" aria-hidden="true">
        {label}
      </span>
    </div>
  )
}
