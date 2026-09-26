import 'i18next'
import type commonJa from './locales/ja/common.json'
import type authJa from './locales/ja/auth.json'
import type calendarJa from './locales/ja/calendar.json'
import type eventJa from './locales/ja/event.json'

declare module 'i18next' {
  interface CustomTypeOptions {
    defaultNS: 'common'
    resources: {
      common: typeof commonJa
      auth: typeof authJa
      calendar: typeof calendarJa
      event: typeof eventJa
    }
  }
}
