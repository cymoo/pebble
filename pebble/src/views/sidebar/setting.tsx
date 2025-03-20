import { ComponentProps, useState } from 'react'

import { cx } from '@/utils/css.ts'

import { Button } from '@/components/button.tsx'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogHeading,
  DialogTrigger,
} from '@/components/dialog.tsx'
import { RadioButton } from '@/components/radio-button.tsx'
import { T, t, useLang } from '@/components/translation.tsx'

import { useIdleTimeout } from '@/views/auth/hooks.tsx'

export function SettingDialog({ className, ...props }: ComponentProps<typeof Button>) {
  return (
    <Dialog>
      <DialogTrigger asChild={true}>
        <Button
          className={cx('text-foreground/85 justify-start! ring-inset text-base!', className)}
          variant="ghost"
          {...props}
        >
          <T name="settings" />
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeading>
          <T name="settings" />
        </DialogHeading>
        <DialogDescription className="sr-only">
          <T name="settingsDescription" />
        </DialogDescription>
        <Setting />
        <DialogClose />
      </DialogContent>
    </Dialog>
  )
}

export function Setting(props: ComponentProps<'div'>) {
  const [fs, setFs] = useState(window.localStorage.getItem('baseFontSize') || '16px')
  const { timeout, setTimeout } = useIdleTimeout()

  const { lang, setLang } = useLang()
  const min = t('minute', lang, false)
  const hour = t('hour', lang, false)

  return (
    <div {...props}>
      <h3 className="text-foreground/80 mb-3">
        <T name="fontSize" />
      </h3>
      <RadioButton
        value={fs}
        onChange={(value) => {
          if (!value) return
          setFs(value)
          document.documentElement.style.fontSize = value
          window.localStorage.setItem('baseFontSize', value)
        }}
        options={[
          { label: <T name="large" />, value: '18px' },
          { label: <T name="medium" />, value: '16px' },
          { label: <T name="small" />, value: '14px' },
        ]}
      />
      <h3 className="text-foreground/80 mt-4 mb-3">
        <T name="language" />
      </h3>
      <RadioButton
        value={lang}
        onChange={(value) => {
          if (!value) return
          setLang(value)
        }}
        options={[
          { label: '中文', value: 'zh' },
          { label: 'English', value: 'en' },
        ]}
      />
      <h3 className="text-foreground/80 mt-4 mb-3">
        <T name="logoutWhenInactivity" />
      </h3>
      <RadioButton
        selfToggleable={true}
        value={timeout}
        onChange={(value) => {
          setTimeout(value)
        }}
        options={[
          { label: `5 ${min}`, value: 5 * 60 },
          { label: `30 ${min}`, value: 30 * 60 },
          { label: `6 ${hour}`, value: 6 * 60 * 60 },
        ]}
      />
    </div>
  )
}
