import { MoonIcon, SunIcon } from 'lucide-react'
import { ComponentProps, useEffect } from 'react'
import { create } from 'zustand'

import { cx } from '@/utils/css.ts'

import { Button } from '@/components/button.tsx'

export function ThemeToggle({ className, ...props }: ComponentProps<typeof Button>) {
  const { theme, setTheme } = useTheme()

  useEffect(() => {
    const handler = (event: MediaQueryListEvent) => {
      setTheme(event.matches ? 'dark' : 'light')
    }
    mql.addEventListener('change', handler)
    return () => {
      mql.removeEventListener('change', handler)
    }
  }, [setTheme])

  return (
    <Button
      className={cx('ring-inset hover:bg-transparent', className)}
      variant="ghost"
      aria-label="toggle dark mode"
      onClick={() => {
        const nextTheme = theme === 'light' ? 'dark' : 'light'
        setTheme(nextTheme)
        localStorage.setItem('theme', nextTheme)
      }}
      {...props}
    >
      {theme === 'dark' ? (
        <MoonIcon className="size-6 align-middle text-yellow-400" />
      ) : (
        <SunIcon className="size-6 align-middle text-yellow-500" />
      )}
    </Button>
  )
}

const mql = window.matchMedia('(prefers-color-scheme: dark)')

const initTheme = () => {
  const isSysDark = mql.matches
  const userTheme = localStorage.getItem('theme')

  document.documentElement.dataset.theme = userTheme ? userTheme : isSysDark ? 'dark' : 'light'
}

export const useTheme = create<{
  theme: 'light' | 'dark'
  setTheme: (theme: 'light' | 'dark') => void
}>((set) => {
  initTheme()

  return {
    theme: (document.documentElement.dataset.theme || 'light') as 'light' | 'dark',
    setTheme: (theme) => {
      set({ theme })
      document.documentElement.dataset.theme = theme
    },
  }
})
