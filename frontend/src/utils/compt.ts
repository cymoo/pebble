import { debounce } from './func'

/**
 * usage in css: `min-height: calc(var(--vh) * 100);`
 */
export function calcViewportUnit() {
  const calc = debounce(() => {
    // First we get the viewport height and multiply it by 1% to get a value for a vh unit
    const vh = window.innerHeight * 0.01
    // Then we set the value in the --vh custom property to the root of the document
    document.documentElement.style.setProperty('--vh', `${vh.toString()}px`)

    const vw = window.innerWidth * 0.01
    document.documentElement.style.setProperty('--vw', `${vw.toString()}px`)
  })

  // initial calculation
  calc()

  // re-calculate on resize
  window.addEventListener('resize', calcViewportUnit)

  // re-calculate on device orientation change
  window.addEventListener('orientationchange', calcViewportUnit)
}
