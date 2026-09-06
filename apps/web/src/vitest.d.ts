import type { TestingLibraryMatchers } from '@testing-library/jest-dom/matchers'

declare module 'vitest' {
  interface Matchers<R, T extends Element | null | undefined = Element>
    extends TestingLibraryMatchers<R, T> {}
}
