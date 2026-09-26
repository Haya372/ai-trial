import commonJa from './locales/ja/common.json'
import authJa from './locales/ja/auth.json'
import calendarJa from './locales/ja/calendar.json'
import eventJa from './locales/ja/event.json'
import commonEn from './locales/en/common.json'
import authEn from './locales/en/auth.json'
import calendarEn from './locales/en/calendar.json'
import eventEn from './locales/en/event.json'

export const resources = {
  ja: { common: commonJa, auth: authJa, calendar: calendarJa, event: eventJa },
  en: { common: commonEn, auth: authEn, calendar: calendarEn, event: eventEn },
} as const
