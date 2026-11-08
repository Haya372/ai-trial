---
name: write-document
description: リポジトリにコミットするドキュメント（ガイドライン・スキル・AGENTS.md・テンプレート・PRD・仕様書・ADR等）を新規作成・更新するとき。「ドキュメントを書く」「ドキュメントを更新する」「ガイドラインに追記する」「スキルを修正する」のような発言や、他のスキルの手順でドキュメントを編集するときに呼び出す。
---

# ドキュメント記述スキル

## 目的

`docs/guidelines/document-writing-guidelines.md` に沿って、読み手がその場で取るべき行動だけが書かれたドキュメントを作成・更新する。

---

## 手順

### ステップ 1: ガイドラインと既存ドキュメントを読む

- `docs/guidelines/document-writing-guidelines.md` を読む
- 書こうとしている内容が既存のドキュメント（`docs/guidelines/`・該当スキル・テンプレート）にないか確認する。ある場合は、その箇所に追記するか参照を置く

### ステップ 2: 書く

- 書く場所ごとに、読み手がその場で取る行動を書く
- テンプレートがあるドキュメント（PRD・仕様書・ADR等）はテンプレートに沿って書く

### ステップ 3: 書いた差分を見直す

`git diff` で追加・変更した行を読み、ガイドラインの「書かない内容」に当たる記述を削る。

### ステップ 4: lintを実行する

```bash
mise exec -- markdownlint-cli2 <変更したファイル>
```

---

## チェックリスト

- [ ] `docs/guidelines/document-writing-guidelines.md` を読んだか
- [ ] 既存ドキュメントと内容が重複していないか
- [ ] 差分を見直し、「書かない内容」に当たる記述を削ったか
- [ ] markdownlintの指摘が0件か

## 参照ドキュメント

- `docs/guidelines/document-writing-guidelines.md` — ドキュメント記述ルール
- `docs/guidelines/document-review-criteria.md` — 書き終えた後のレビュー観点
