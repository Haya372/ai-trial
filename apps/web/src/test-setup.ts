import { cleanup } from '@testing-library/react'
import * as matchers from '@testing-library/jest-dom/matchers'
import { afterEach, expect } from 'vitest'
import './i18n/test-init'

expect.extend(matchers)
afterEach(cleanup)
