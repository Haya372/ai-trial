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
import { useTranslation } from 'react-i18next'
import { useLoginForm } from '../hooks/useLoginForm'

interface LoginPageProps {
  onSuccess?: () => void
}

export default function LoginPage({ onSuccess }: LoginPageProps) {
  const { t } = useTranslation('auth')
  const { form, onSubmit } = useLoginForm(onSuccess)

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className="flex flex-col gap-4 max-w-sm mx-auto mt-16"
      >
        <Text variant="h2">{t('login.title')}</Text>
        <FormField
          control={form.control}
          name="email"
          render={({ field, fieldState }) => (
            <FormItem>
              <FormLabel>{t('fields.email')}</FormLabel>
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
              <FormLabel>{t('fields.password')}</FormLabel>
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
          {form.formState.isSubmitting
            ? t('login.submitButtonLoading')
            : t('login.submitButton')}
        </Button>
        <Link
          to="/signup"
          className="text-sm text-primary underline-offset-4 hover:underline"
        >
          {t('login.signupLink')}
        </Link>
      </form>
    </Form>
  )
}
