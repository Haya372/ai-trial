# i18n（国際化）基盤

`apps/web` の国際化（i18n）基盤のドキュメント。採用ライブラリの選定理由は [ADR-029](../../../../docs/adr/ADR-029-i18n-library.md) を参照。

## 概要

- ライブラリ: `i18next` + `react-i18next` + `i18next-browser-languagedetector`
- 翻訳リソースは `locales/{lang}/{namespace}.json` に配置
- 言語設定は `localStorage['i18nextLng']` に永続化される
- フォールバック言語: `ja`（対応言語: `ja` / `en`）

## ファイル構成

```
src/i18n/
├── config.ts          # i18next の初期化（副作用モジュール）
├── resources.ts       # 全ロケールのJSONを集約
├── i18n.d.ts          # TypeScript 型augmentation（キー補完を有効化）
├── test-init.ts       # テスト用初期化（test-setup.ts から import）
├── README.md          # このファイル
└── locales/
    ├── ja/            # 日本語リソース
    │   ├── common.json
    │   ├── auth.json
    │   ├── calendar.json
    │   └── event.json
    └── en/            # 英語リソース（ja と同じキー構成）
        ├── common.json
        ├── auth.json
        ├── calendar.json
        └── event.json
```

## 名前空間

| 名前空間 | 対応するフィーチャー | 用途 |
|---|---|---|
| `common` | 共通 | ボタンラベル・共通エラー等、フィーチャー横断の汎用文言 |
| `auth` | `features/auth/` | ログイン・サインアップ等 |
| `calendar` | `features/calendar/` | カレンダー表示・操作 |
| `event` | `features/event/` | イベント登録・編集 |

新しいフィーチャーを追加する場合は、`features/` 配下のフィーチャー名と同名の名前空間を追加してください。

## 新しい文言の追加手順

1. 適切な名前空間を選ぶ（フィーチャーと同名、またはフィーチャー横断なら `common`）
2. `locales/ja/{namespace}.json` にキーと日本語文言を追加する

   ```json
   {
     "form": {
       "submit": "送信"
     }
   }
   ```

3. `locales/en/{namespace}.json` にも**同一キー**を追加する（英語訳が未確定の場合は日本語をそのまま入れて後日置換する。空文字列は不可）

   ```json
   {
     "form": {
       "submit": "Submit"
     }
   }
   ```

4. コンポーネントで呼び出す

   ```tsx
   import { useTranslation } from 'react-i18next'

   const { t } = useTranslation('auth')

   // 同じ名前空間内のキー
   t('form.submit')

   // 別の名前空間のキーを参照する場合
   t('common:save')
   ```

### キー命名規則

- `lowerCamelCase` でドット区切りの階層構造を使う（例: `form.submitButton`、`validation.emailRequired`）
- ボタンラベルは体言止め（例: `save` / `cancel` / `delete`）
- 動的補間を伴うキーに特別な接尾辞は付けない（`{{var}}` 記法で自然に判別できる）

## 補間の書き方

翻訳リソースに `{{変数名}}` 形式で埋め込み、`t()` 呼び出し時に値を渡す。

```json
{
  "greeting": "こんにちは、{{name}}さん"
}
```

```tsx
t('greeting', { name: '田中' })  // → "こんにちは、田中さん"
```

## 言語切り替え

コンポーネントから直接 `i18n.changeLanguage()` を呼ぶ。

```tsx
import { useTranslation } from 'react-i18next'

const { i18n } = useTranslation()

// 言語を英語に切り替え（localStorageに永続化される）
await i18n.changeLanguage('en')
```

言語選択UIは未実装（別Issueで対応予定）。ブラウザの devtools コンソールから確認する場合は:

```js
localStorage.setItem('i18nextLng', 'en')
// ページリロードすると英語表示になる
```

## Zodスキーマでの利用（パターン A 推奨）

Zodスキーマ内のバリデーションメッセージを i18n 化する場合、フォームフック内で `t` 関数を使ってスキーマを構築する。

```tsx
// features/auth/hooks/useLoginForm.ts（Issue #207 で実装予定）
const { t } = useTranslation('auth')
const schema = useMemo(
  () =>
    z.object({
      email: z.string().min(1, t('validation.emailRequired')).email(t('validation.emailInvalid')),
    }),
  [t],
)
```

詳細な実装パターンは Issue #207 で確立予定。

## 将来対応

- **言語切り替えUI**: メニュー等のUIは未実装。別Issueで追加予定
- **`ja`/`en` のキー parity チェック**: `en` の翻訳が本格的に追加され始めたタイミングで、parity検査スクリプトを別Issueとして起票する予定
- **動的リソースロード**: 現在は静的importでリソースを同梱。翻訳量が増えた場合は `i18next-http-backend` への移行を検討する
