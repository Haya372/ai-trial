import i18n from 'i18next'
import { afterEach } from 'vitest'
import './config'

// jsdom's navigator.language defaults to 'en-US', which the language detector
// picks up before any test runs, overriding the app's 'ja' fallback. Force
// 'ja' up front so tests start from the same language regardless of jsdom.
await i18n.changeLanguage('ja')

// Reset language to 'ja' after each test so changeLanguage() calls don't bleed across test files
afterEach(async () => {
  await i18n.changeLanguage('ja')
})
