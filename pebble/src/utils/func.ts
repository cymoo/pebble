interface DebouncedFunction<T extends (...args: never[]) => unknown> {
  (...args: Parameters<T>): void
  cancel: () => void
}

export function debounce<T extends (...args: never[]) => unknown>(
  callback: T,
  timeout = 50,
): DebouncedFunction<T> {
  let timer: ReturnType<typeof setTimeout> | undefined

  const debounced = (...args: Parameters<T>) => {
    clearTimeout(timer)
    timer = setTimeout(() => callback(...args), timeout)
  }

  debounced.cancel = () => {
    clearTimeout(timer)
    timer = undefined
  }

  return debounced as DebouncedFunction<T>
}

export function noop() {
  /**/
}

export function delay(ms: number): Promise<undefined> {
  return new Promise((resolve) => {
    setTimeout(resolve, ms)
  })
}
