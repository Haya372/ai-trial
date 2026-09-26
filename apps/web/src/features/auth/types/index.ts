import type { TFunction } from 'i18next'
import { z } from 'zod'

export function createLoginSchema(t: TFunction<'auth'>) {
  return z.object({
    email: z
      .string()
      .min(1, t('validation.emailRequired'))
      .email(t('validation.emailInvalid')),
    password: z.string().min(1, t('validation.passwordRequired')),
  })
}

export function createSignupSchema(t: TFunction<'auth'>) {
  return z.object({
    email: z
      .string()
      .min(1, t('validation.emailRequired'))
      .email(t('validation.emailInvalid')),
    password: z
      .string()
      .min(8, t('validation.passwordMinLength'))
      .max(128, t('validation.passwordMaxLength')),
    displayName: z
      .string()
      .max(50, t('validation.displayNameMaxLength'))
      .optional(),
  })
}

export type LoginFormValues = z.infer<ReturnType<typeof createLoginSchema>>
export type SignupFormValues = z.infer<ReturnType<typeof createSignupSchema>>
