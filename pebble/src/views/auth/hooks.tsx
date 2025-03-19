import { useCallback } from 'react'
import { Location, useLocation, useNavigate } from 'react-router'
import useSWRMutation from 'swr/mutation'
import { create } from 'zustand'

import { removeCookie, setCookie } from '@/utils/cookie.ts'
import { useIdle } from '@/utils/hooks/use-idle.ts'

import { postActions as actions } from '@/views/actions.ts'

import { LOGIN, fetcher } from '@/api.ts'
import { AppError } from '@/error.ts'

export function useLogin() {
  const { trigger, error, isMutating } = useSWRMutation<
    { token: string }, // Data
    AppError, // Error
    string, // Key
    string // ExtraArg
  >(LOGIN, (url: string, { arg }: { arg: string }) => {
    return fetcher(url, { password: arg })
  })

  const navigate = useNavigate()
  const location = useLocation() as Location<{ from?: { pathname: string } } | undefined>

  const from = location.state?.from?.pathname || '/'

  const handleLogin = async (password: string) => {
    if (!password) {
      return
    }
    await trigger(password)
    localStorage.setItem('token', password)
    // NOTE: Cookie is used to utilize nginx `auth_request`
    setCookie('token', password, -1)

    void navigate(from, { replace: true })
  }

  return {
    error,
    isLoading: isMutating,
    handleLogin,
  }
}

export function useLogout() {
  const navigate = useNavigate()

  return useCallback(() => {
    localStorage.removeItem('token')
    removeCookie('token')

    void actions.clearCaches().then(() => {
      void navigate('/login')
    })
  }, [navigate])
}

export function useLogoutWhenInactive() {
  const timeout = useIdleTimeout((state) => state.timeout)
  const logout = useLogout()
  useIdle(timeout, () => {
    logout()
  })
}

export const useIdleTimeout = create<{
  timeout: number | null
  setTimeout: (timeout: number | null) => void
}>((set) => ({
  timeout: Number(window.localStorage.getItem('idleTimeout')) || null,
  setTimeout: (timeout: number | null) => {
    set({ timeout })
    if (timeout === null) {
      window.localStorage.removeItem('idleTimeout')
    } else {
      window.localStorage.setItem('idleTimeout', String(timeout))
    }
  },
}))
