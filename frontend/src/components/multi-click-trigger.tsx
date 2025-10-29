import { ComponentProps, useRef, useState } from 'react'

import { Button } from '@/components/button.tsx'

interface MultiClickTriggerProps extends ComponentProps<typeof Button> {
  /**
   * Number of clicks required to trigger the action
   * @default 10
   */
  clickCount?: number

  /**
   * Maximum time interval between clicks in milliseconds
   * @default 500
   */
  timeInterval?: number

  /**
   * Callback function triggered when the required clicks are completed
   */
  onTrigger?: () => void
}

export function MultiClickTrigger({
  clickCount = 10,
  timeInterval = 1000,
  onTrigger,
  ...props
}: MultiClickTriggerProps) {
  const [clicks, setClicks] = useState<number>(0)
  const lastClickRef = useRef<number>(0)

  const handleClick = (): void => {
    const now = Date.now()
    const timeSinceLastClick = now - lastClickRef.current

    // Update the timestamp of the last click
    lastClickRef.current = now

    // If the time since last click exceeds the interval, reset counter to 1
    if (timeSinceLastClick > timeInterval) {
      setClicks(1)
      return
    }

    // If within valid time interval, increment click count
    const newClickCount = clicks + 1

    // Check if target click count has been reached
    if (newClickCount >= clickCount) {
      // Target reached, trigger callback and reset
      onTrigger?.()
      setClicks(0)
    } else {
      // Not yet reached target, update counter
      setClicks(newClickCount)
    }
  }

  return <Button {...props} onClick={handleClick} />
}
