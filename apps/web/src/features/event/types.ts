import { z } from 'zod'

export const eventFormSchema = z
  .object({
    title: z.string().min(1, 'タイトルを入力してください'),
    startAt: z.string().min(1, '開始日時を入力してください'),
    endAt: z.string().min(1, '終了日時を入力してください'),
    description: z.string().optional(),
    location: z.string().optional(),
    url: z.string().optional(),
  })
  .refine((data) => new Date(data.endAt) > new Date(data.startAt), {
    message: '終了日時は開始日時より後に設定してください',
    path: ['endAt'],
  })

export type EventFormValues = z.infer<typeof eventFormSchema>
export type EventFormMode = 'create' | 'edit'
