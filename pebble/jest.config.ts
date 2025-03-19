import type { Config } from '@jest/types'
const config: Config.InitialOptions = {
  verbose: true,
  transform: {
    '^.+\\.tsx?$': 'ts-jest',
  },
  testEnvironment: 'jsdom',
  roots: ['./tests'],
}
export default config
