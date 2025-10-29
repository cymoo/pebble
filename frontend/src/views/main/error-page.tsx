import { ComponentProps, ReactNode } from 'react'

import { CenteredContainer } from '@/views/layout/layout.tsx'

interface ErrorPageProps extends ComponentProps<'div'> {
  title: string
  description: ReactNode
  extra?: ReactNode
}

export function ErrorPage({ title, description, extra, ...props }: ErrorPageProps) {
  return (
    <CenteredContainer title={title} {...props}>
      <p className="text-muted-foreground mb-4 text-lg">{description}</p>
      <div className="text-center">{extra}</div>
    </CenteredContainer>
  )
}
