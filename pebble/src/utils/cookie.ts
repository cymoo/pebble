/**
 * Sets a cookie with the specified name, value, and optional expiration duration days.
 * @param name - The name of the cookie.
 * @param value - The value to be stored in the cookie.
 * @param expired_days - Optional. The number of days until the cookie expires.
 *                       If not provided, the cookie will not persist after the browser is closed.
 *                       If set to -1, the cookie will persist forever.
 */
export function setCookie(name: string, value?: string, expired_days?: number) {
  let expires = ''

  if (expired_days) {
    if (expired_days === -1) {
      expired_days = 365 * 10
    }
    const date = new Date()
    date.setDate(date.getDate() + expired_days)
    expires = '; expires=' + date.toUTCString()
  }

  document.cookie = name + '=' + (value ?? '') + expires + '; path=/'
}

export function removeCookie(name: string) {
  document.cookie = name + '=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;'
}
