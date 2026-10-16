import { Button, toast } from '@repo/ui'
import { useQueryClient } from '@tanstack/react-query'
import { useNavigate } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { logout } from '../../../api/generated'
import { useAuthStore } from '../../../stores/auth'

export default function LogoutButton() {
  const { t } = useTranslation('auth')
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const clearUser = useAuthStore((s) => s.clearUser)

  const handleLogout = async () => {
    try {
      const res = await logout()
      // 401 はサーバー側でセッションが既に無効なので、ローカル状態も破棄してよい
      if (res.status !== 204 && res.status !== 401) {
        toast.error(t('logout.errors.fallback'))
        return
      }
      clearUser()
      // 同じタブで別ユーザーがログインしたとき、前ユーザーのキャッシュが見えないようにする
      queryClient.clear()
      await navigate({ to: '/login' })
    } catch {
      toast.error(t('logout.errors.fallback'))
    }
  }

  return (
    <Button variant="secondary" onClick={handleLogout}>
      {t('logout.button')}
    </Button>
  )
}
