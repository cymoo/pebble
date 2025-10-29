import { ComponentProps, ReactElement, useEffect } from 'react'

import { cx } from '@/utils/css.ts'

import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerHeading,
  DrawerTrigger,
} from '@/components/drawer.tsx'

import { CONTENT_WIDTH, SIDEBAR_WIDTH } from '@/constants.ts'

import { MainHeader } from './header.tsx'
import { useIsSmallDevice } from './hooks.tsx'

declare global {
  interface Window {
    toggleSidebar: () => void
  }
}

export const PrimaryLayout = ({ side, main }: { side: ReactElement; main: ReactElement }) => {
  const sm = useIsSmallDevice()

  useEffect(() => {
    // ah, it's not that react
    window.toggleSidebar = () => {
      document.getElementById('sidebar-toggle')?.click()
    }
  }, [])

  let sidebar = (
    <aside
      className={cx('scrollbar-none vh-full flex-none overflow-x-hidden overflow-y-auto px-4 py-3')}
      style={{ width: sm ? '100%' : SIDEBAR_WIDTH }}
    >
      {side}
    </aside>
  )

  if (sm) {
    sidebar = (
      <Drawer
        outsidePress={(event) => {
          return (event.target as HTMLElement).classList.contains('sidebar-overlay')
        }}
      >
        <DrawerTrigger className="sr-only" id="sidebar-toggle">
          toggle sidebar
        </DrawerTrigger>
        <DrawerContent
          overlayClassName="sidebar-overlay"
          className="w-[80vw] max-w-sm! p-0!"
          side="left"
        >
          <DrawerHeading className="sr-only">side bar</DrawerHeading>
          <DrawerClose className="sr-only" />
          {sidebar}
        </DrawerContent>
      </Drawer>
    )
  }

  return (
    <div className={cx('mx-auto flex', sm ? 'h-auto w-full' : 'h-screen w-fit')}>
      {sidebar}
      <main
        className={cx('relative flex flex-col px-4 py-3', { 'pt-15': sm })}
        style={{ width: sm ? '100%' : CONTENT_WIDTH }}
      >
        <MainHeader className="flex-none" />
        {main}
      </main>
    </div>
  )
}

export const CenteredContainer = ({
  title,
  className,
  children,
  ...props
}: { title: string } & ComponentProps<'div'>) => {
  return (
    <div className={cx('flex flex-col items-center px-3 pt-[10vh]', className)} {...props}>
      <title>{title}</title>
      <h1 className="mb-5 text-2xl font-semibold">{title}</h1>
      {children}
    </div>
  )
}

export const ContentContainer = ({
  className,
  style,
  children,
  ...props
}: ComponentProps<'div'>) => {
  return (
    <div
      className={cx('mx-auto', className)}
      style={{ maxWidth: CONTENT_WIDTH, ...style }}
      {...props}
    >
      {children}
    </div>
  )
}
