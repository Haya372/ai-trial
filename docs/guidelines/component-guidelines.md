# コンポーネント実装ガイドライン

ADR-015でshadcn/ui（Tailwind CSS + Radix UI）を採用した。このドキュメントはUIコンポーネントを実装・利用するときの規約を定義する。

## 基本原則

### variantで制御し、classNameを受け取らない

コンポーネントの見た目はすべて `variant` / `size` などのProps経由で制御する。`className` Propを外部に公開しない。

```tsx
// OK: variantで見た目を制御する
<Button variant="destructive" size="sm">削除</Button>

// NG: classNameで上書きする
<Button className="bg-red-500 text-xs">削除</Button>
```

**理由**: `className` を許容すると、デザインシステムの一貫性が任意の場所から崩せる。想定外のスタイルの混入を防ぐために、コンポーネントが自身の見た目を完全に管理する。

### variantの実装には `cva` を使う

[class-variance-authority](https://cva.style/)（`cva`）でvariantとTailwindクラスの対応を宣言する。

```tsx
import { cva, type VariantProps } from "class-variance-authority";

const buttonVariants = cva(
  // ベースクラス（全variantに共通）
  "inline-flex items-center justify-center rounded-md font-medium transition-colors focus-visible:outline-none disabled:pointer-events-none disabled:opacity-50",
  {
    variants: {
      variant: {
        primary:     "bg-primary text-primary-foreground hover:bg-primary/90",
        secondary:   "bg-secondary text-secondary-foreground hover:bg-secondary/80",
        destructive: "bg-destructive text-destructive-foreground hover:bg-destructive/90",
        outline:     "border border-input bg-background hover:bg-accent",
        ghost:       "hover:bg-accent hover:text-accent-foreground",
        link:        "text-primary underline-offset-4 hover:underline",
      },
      size: {
        sm: "h-8 px-3 text-xs",
        md: "h-9 px-4 text-sm",
        lg: "h-11 px-6 text-base",
      },
    },
    defaultVariants: {
      variant: "primary",
      size: "md",
    },
  }
);

type ButtonProps = React.ButtonHTMLAttributes<HTMLButtonElement> &
  VariantProps<typeof buttonVariants>;

function Button({ variant, size, ...props }: ButtonProps) {
  return (
    <button className={buttonVariants({ variant, size })} {...props} />
  );
}
```

---

## intentバリアント

コンポーネントの意図（アクションの重要度・性質）を表す。`variant` Propで指定する。

| variant | 意味 | 使う場面 |
|---|---|---|
| `primary` | 最も重要なアクション | 予定の保存、フォームのSubmit、確認ダイアログのOK |
| `secondary` | 補助的なアクション | フォームのリセット、「後で」ボタン |
| `destructive` | 取り消せない危険な操作 | 予定の削除、アカウント削除 |
| `outline` | 低優先度のアクション | キャンセル、戻る、副次的なCTA |
| `ghost` | 極力目立たせないアクション | ツールバーのアイコンボタン、ナビゲーション項目 |
| `link` | インラインのテキストリンク的な操作 | 文章中の操作、設定リンク |

### 使い分けの判断基準

**1画面に `primary` は1つを原則とする**。複数の主要アクションが並ぶ場合は重要度を決め、2番目以降は `secondary` または `outline` に落とす。

```tsx
// OK: 主アクションは1つ、副次は outline
<Button variant="primary">保存</Button>
<Button variant="outline">キャンセル</Button>

// NG: primary が並ぶ
<Button variant="primary">保存</Button>
<Button variant="primary">下書き保存</Button>
```

**`destructive` は確認ステップとセットにする**。単独で配置せず、確認ダイアログ内の最終確認ボタンとして使う。

```tsx
// OK: 確認ダイアログの中で使う
<Dialog>
  <DialogTrigger asChild>
    <Button variant="outline">予定を削除</Button>   {/* 起点は outline */}
  </DialogTrigger>
  <DialogContent>
    <DialogFooter>
      <Button variant="outline">キャンセル</Button>
      <Button variant="destructive">削除する</Button>  {/* 最終確認で destructive */}
    </DialogFooter>
  </DialogContent>
</Dialog>
```

**`ghost` はアイコンボタンのデフォルト**。テキストラベルなしのアイコンボタンはすべて `ghost` にする。

```tsx
<Button variant="ghost" size="sm" aria-label="閉じる">
  <XIcon />
</Button>
```

---

## sizeバリアント

コンポーネントの物理的な大きさを表す。`size` Propで指定する。

| size | 用途 |
|---|---|
| `sm` | ツールバー、コンパクトなリスト行、バッジに近い操作 |
| `md` | 大多数のボタン（デフォルト） |
| `lg` | ページ単位の主要CTA、モーダルのメインアクション |

`size` は `variant` と独立して組み合わせられる。

```tsx
<Button variant="primary" size="lg">新しい予定を作成</Button>
<Button variant="ghost" size="sm" aria-label="編集"><PencilIcon /></Button>
```

---

## 状態の扱い

`disabled` / `loading` はHTML標準属性またはPropで表現する。variantには含めない。

```tsx
// disabled: HTML属性で渡す
<Button variant="primary" disabled>保存中...</Button>

// loading: ローディング状態を示すPropを別途定義する
<Button variant="primary" loading>
  <Spinner />
  保存中...
</Button>
```

---

## 新しいvariantを追加するとき

1. **既存variantで代替できないか先に確認する**。見た目だけ微妙に違うvariantを増やさない
2. 追加する場合はこのドキュメントの対応表を更新し、用途の説明を明記する
3. shadcn/uiのコンポーネントソース（`src/components/ui/`）の `cva` 定義に追加する

```tsx
// src/components/ui/button.tsx
const buttonVariants = cva("...", {
  variants: {
    variant: {
      // 既存 ...
      warning: "bg-yellow-500 text-white hover:bg-yellow-500/90",  // 追加例
    },
  },
});
```

---

## 参照

- [ADR-015: フロントエンドコンポーネントライブラリ](../adr/ADR-015-frontend-component-library.md)
- [設計ガイドライン](design-guidelines.md)
