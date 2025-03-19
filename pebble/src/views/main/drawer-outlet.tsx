import { useLayoutEffect, useState } from 'react'
import { Outlet, useLocation } from 'react-router'

import { Drawer, DrawerContent } from '@/components/drawer.tsx'
import { useStableNavigate } from '@/components/router.tsx'

import { useIsSmallDevice } from '@/views/layout/hooks.tsx'

export const DrawerOutlet = () => {
  const location = useLocation()
  const navigate = useStableNavigate()

  const shouldCloseDrawer = (location.state as { isFirstLayer?: boolean } | null)?.isFirstLayer

  const sm = useIsSmallDevice()

  // NOTE: Allows the drawer to have the time to start the animation
  const [visible, setVisible] = useState(false)
  useLayoutEffect(() => {
    setVisible(true)
  }, [])

  return (
    <Drawer
      open={visible}
      onOpenChange={setVisible}
      outsidePress={(event) => {
        // Try to close the drawer when clicking overlay
        return (event.target as HTMLElement).classList.contains('post-overlay')
      }}
      beforeDismiss={(close) => {
        if (shouldCloseDrawer) {
          // Start the transition animation before the drawer closes
          close()
        } else {
          // Do not close the drawer
          void navigate(-1)
        }
      }}
    >
      <DrawerContent
        overlayClassName="post-overlay"
        className="scrollbar-none overflow-y-auto p-0!"
        style={{ width: sm ? '90vw' : '630px' }}
        afterLeave={() => {
          // Modify the URL after the animation completes
          void navigate(-1)
        }}
      >
        <Outlet />
      </DrawerContent>
    </Drawer>
  )
}
