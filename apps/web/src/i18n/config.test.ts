import { afterEach, describe, expect, it } from 'vitest'
import i18n from 'i18next'
import './config'

describe('i18n smoke test', () => {
  afterEach(async () => {
    await i18n.changeLanguage('ja')
  })

  it('returns "English" for common:languageName when language is en', async () => {
    await i18n.changeLanguage('en')
    expect(i18n.t('common:languageName')).toBe('English')
  })

  it('returns "日本語" for common:languageName when language is ja', async () => {
    await i18n.changeLanguage('ja')
    expect(i18n.t('common:languageName')).toBe('日本語')
  })
})
