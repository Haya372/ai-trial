import i18n from 'i18next'
import { afterEach } from 'vitest'
import './config'

// Reset language to 'ja' after each test so changeLanguage() calls don't bleed across test files
afterEach(async () => {
  await i18n.changeLanguage('ja')
})
