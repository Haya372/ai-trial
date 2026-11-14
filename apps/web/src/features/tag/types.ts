import { TAG_COLOR_KEYS } from '@repo/ui'
import type { TFunction } from 'i18next'
import { z } from 'zod'
import type { TagResponse } from '../../api/generated'

export const TAG_LIMIT = 50
const TAG_NAME_MAX_LENGTH = 20

export function createTagFormSchema(t: TFunction<'tag'>) {
  return z.object({
    name: z
      .string()
      .trim()
      .min(1, t('validation.nameRequired'))
      // Counts code points like the backend's []rune check, so emoji count as one character.
      .refine((name) => [...name].length <= TAG_NAME_MAX_LENGTH, {
        message: t('validation.nameTooLong', { max: TAG_NAME_MAX_LENGTH }),
      }),
    color: z.enum(TAG_COLOR_KEYS),
  })
}

export type TagFormInput = z.input<ReturnType<typeof createTagFormSchema>>
export type TagFormValues = z.output<ReturnType<typeof createTagFormSchema>>

export type TagEditorTarget =
  | { mode: 'create' }
  | { mode: 'edit'; tag: TagResponse }
