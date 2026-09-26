import {
  Button,
  Form,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  Input,
  Text,
} from '@repo/ui'
import { Link } from '@tanstack/react-router'
import { useLoginForm } from '../hooks/useLoginForm'

interface LoginPageProps {
  onSuccess?: () => void
}

export default function LoginPage({ onSuccess }: LoginPageProps) {
  const { form, onSubmit } = useLoginForm(onSuccess)

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className="flex flex-col gap-4 max-w-sm mx-auto mt-16"
      >
        <Text variant="h2">ログイン</Text>
        <FormField
          control={form.control}
          name="email"
          render={({ field, fieldState }) => (
            <FormItem>
              <FormLabel>メールアドレス</FormLabel>
              <Input
                id={field.name}
                type="email"
                state={fieldState.error ? 'error' : 'default'}
                aria-invalid={!!fieldState.error}
                {...field}
              />
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="password"
          render={({ field, fieldState }) => (
            <FormItem>
              <FormLabel>パスワード</FormLabel>
              <Input
                id={field.name}
                type="password"
                state={fieldState.error ? 'error' : 'default'}
                aria-invalid={!!fieldState.error}
                {...field}
              />
              <FormMessage />
            </FormItem>
          )}
        />
        <Button
          type="submit"
          variant="primary"
          disabled={form.formState.isSubmitting}
        >
          {form.formState.isSubmitting ? 'ログイン中…' : 'ログイン'}
        </Button>
        <Link
          to="/signup"
          className="text-sm text-primary underline-offset-4 hover:underline"
        >
          アカウントをお持ちでない方はこちら
        </Link>
      </form>
    </Form>
  )
}
