import { KeyIcon } from 'lucide-react'
import { ComponentProps, useEffect, useRef } from 'react'

import { Input } from '@/components/input.tsx'
import { Spinner } from '@/components/spinner.tsx'
import { T, t, useLang } from '@/components/translation.tsx'

import { CenteredContainer } from '../layout/layout.tsx'
import { useLogin } from './hooks.tsx'

export function Login(props: ComponentProps<'div'>) {
  const { error, isLoading, handleLogin } = useLogin()
  const { lang } = useLang()

  // focus input again when password is wrong
  const inputRef = useRef<HTMLInputElement>(null)
  useEffect(() => {
    if (error) inputRef.current?.focus()
  }, [error])

  return (
    <CenteredContainer title="Login" {...props}>
      <div className="flex items-center justify-between space-x-5">
        <Input
          ref={inputRef}
          type="password"
          className="w-[calc(max(20vw,250px))]!"
          prefix={<KeyIcon className="size-5" />}
          autoFocus
          // https://developer.mozilla.org/zh-CN/docs/Web/Security/Securing_your_site/Turning_off_form_autocompletion
          autoComplete="new-password"
          placeholder={t('password', lang)}
          disabled={isLoading}
          onKeyUp={(event) => {
            if (event.key.toLowerCase() === 'enter') {
              void handleLogin((event.target as HTMLInputElement).value.trim())
            }
          }}
          onBlur={(event) => {
            void handleLogin(event.target.value.trim())
          }}
        />
      </div>
      {(isLoading || error) && (
        <div className="text-destructive mt-3 text-center">
          {isLoading ? (
            <Spinner className="text-primary" />
          ) : error?.code === 500 ? (
            <T name="serverError" />
          ) : error?.code === 429 ? (
            <T name="tooManyFails" />
          ) : (
            <T name="wrongPassword" />
          )}
        </div>
      )}
    </CenteredContainer>
  )
}
