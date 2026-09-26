import i18n from 'i18next'
import { beforeEach } from 'vitest'
import './config'

// Force language to 'ja' before each test: LanguageDetector can pick up 'en'
// from jsdom's default navigator.language, and changeLanguage() calls in one
// test must not bleed into the next.
beforeEach(async () => {
  await i18n.changeLanguage('ja')
})
