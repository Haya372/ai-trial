import { create } from 'zustand'
import {
  CALENDAR_VIEW,
  type CalendarView,
} from '../features/calendar/constants'

export interface CalendarState {
  view: CalendarView
  currentDate: Date
  setView: (view: CalendarView) => void
  setCurrentDate: (date: Date) => void
}

export const useCalendarStore = create<CalendarState>((set) => ({
  view: CALENDAR_VIEW.MONTH,
  currentDate: new Date(),
  setView: (view) => set({ view }),
  setCurrentDate: (date) => set({ currentDate: date }),
}))
