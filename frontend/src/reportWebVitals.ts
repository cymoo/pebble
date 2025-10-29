import { Metric } from 'web-vitals'

const reportWebVitals = (onPerfEntry?: (metric: Metric) => void) => {
  if (onPerfEntry) {
    import('web-vitals')
      .then(({ onCLS, onINP, onFCP, onLCP, onTTFB }) => {
        onCLS(onPerfEntry)
        onINP(onPerfEntry)
        onFCP(onPerfEntry)
        onLCP(onPerfEntry)
        onTTFB(onPerfEntry)
      })
      .catch((err: unknown) => {
        console.log('cannot import web-vitals: ', err)
      })
  }
}

export default reportWebVitals
