import { z } from 'zod'

export const loginSchema = z.object({
  email: z
    .string()
    .min(1, 'メールアドレスを入力してください')
    .email('正しいメールアドレスを入力してください'),
  password: z.string().min(1, 'パスワードを入力してください'),
})

export const signupSchema = z.object({
  email: z
    .string()
    .min(1, 'メールアドレスを入力してください')
    .email('正しいメールアドレスを入力してください'),
  password: z
    .string()
    .min(8, 'パスワードは8文字以上で入力してください')
    .max(128, 'パスワードは128文字以内で入力してください'),
  displayName: z
    .string()
    .max(50, '表示名は50文字以内で入力してください')
    .optional(),
})

export type LoginFormValues = z.infer<typeof loginSchema>
export type SignupFormValues = z.infer<typeof signupSchema>
