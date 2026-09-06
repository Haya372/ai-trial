---
name: commit
description: コミットするとき。適切な粒度でコミットを分割し、規約に従ったメッセージを作成する。
---

# コミットスキル

## コミットの粒度

**1コミット = 1論理的変更**

コミットは「なぜ」の単位で分割する。ファイル数の多さは関係ない。

| 良い例 | 悪い例 |
|---|---|
| テスト追加と実装を別コミット | 複数機能をまとめて1コミット |
| リファクタリング専用コミット | 機能追加とバグ修正を混在 |
| 設定変更専用コミット | 「WIP」や「修正」だけのメッセージ |

分割の判断基準: **コミットをrevertしたとき、他の変更に影響が出ないか？** 影響が出るなら分割する。

## コミットメッセージ規約

### フォーマット

```
<type>: <summary>

[body]
```

### type

| type | 用途 |
|---|---|
| `feat` | 新機能 |
| `fix` | バグ修正 |
| `docs` | ドキュメントのみの変更 |
| `refactor` | 動作を変えないコード整理 |
| `test` | テストの追加・修正 |
| `chore` | ビルド・設定・依存関係などの雑務 |

### summary（1行目）

- **英語・命令形**で書く（例: `Add`, `Fix`, `Remove`）
- **50文字以内**
- 末尾にピリオドをつけない

### body（本文）

- 必要な場合のみ書く（WHYが自明でない場合）
- 1行目との間に空行を1行入れる
- **72文字で折り返す**
- WHATではなくWHYを書く

### 例

```
feat: Add explorer agent for code investigation

Use Haiku model to reduce context consumption during
large-scale code searches.
```

## コマンド

```bash
# ステージング
git add <file>...

# コミット（本文なし）
git commit -m "<type>: <summary>"

# コミット（本文あり）
git commit -m "$(cat <<'EOF'
<type>: <summary>

<body>
EOF
)"
```

## チェックリスト

- [ ] 1コミット1論理的変更になっているか
- [ ] typeが適切か
- [ ] summaryが英語・命令形・50文字以内か
- [ ] WHYが不明な場合は本文にWHYを書いているか
