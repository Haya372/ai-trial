import { create } from 'zustand'

type CalendarView = 'month' | 'week'

export interface CalendarState {
  view: CalendarView
  currentDate: Date
  setView: (view: CalendarView) => void
  setCurrentDate: (date: Date) => void
}

export const useCalendarStore = create<CalendarState>((set) => ({
  view: 'month',
  currentDate: new Date(),
  setView: (view) => set({ view }),
  setCurrentDate: (date) => set({ currentDate: date }),
}))
