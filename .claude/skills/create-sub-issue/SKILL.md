---
name: create-sub-issue
description: 大きすぎるIssueをsub-issueに分割するとき。「Issueが大きい」「sub-issueに分割したい」「Issueを分割する」のような発言で呼び出す。
---

# Sub-issue分割スキル

## 目的

粒度が大きすぎるIssue（多くはPRDから直接起票された機能Issue）を、実装可能な単位のsub-issueに分割する。
分割すべきかどうかの判断自体は `start-issue` スキルが行い、分割が必要と判断された場合にこのスキルを呼び出す。

---

## 手順

### ステップ 1: 対象Issueと関連PRDを読む

```bash
gh issue view <number>
```

- 受け入れ条件を洗い出し、どのレイヤー（DB / ドメイン / API / フロントエンド）にまたがるか整理する
- 関連PRDがあれば読み、目標・非目標・制約を把握する

### ステップ 2: 類似の分割事例を探す

過去に同種の機能で分割した実績があれば、粒度・命名・依存関係の書き方を踏襲する。

```bash
gh issue list --state all --search "in:body 親Issue" --json number,title,body
```

見つかった場合は該当Issue本文（特に「設計メモ」の書き方）を確認し、フォーマットを揃える。

### ステップ 3: コードベースの構成を確認する

対象機能が触れるレイヤー（例: `apps/backend/domain` / `apps/backend/usecase` / `apps/backend/interface` / `apps/backend/db/migrations` / `apps/web/src/features`）をExploreし、既存の類似実装のディレクトリ構造を確認する。

### ステップ 4: レイヤー単位で分割案を作る

分割の粒度は `docs/guidelines/development-flow.md` の「PR分割ガイドライン」にある独立した関心事の単位（DBスキーマ/APIエンドポイント/フロントエンドUI）に合わせる。

**分割順序はADR-001のDependency Rule（依存の向きは常にDomain層へ）に従う。**

- Domain層（エンティティ・リポジトリインターフェース）は何にも依存しないため、まず設計・決定する
- Infrastructure層（DBマイグレーション・リポジトリ実装）はDomain層が定義したインターフェースを満たす実装でしかないため、後で着手する
- 「DBスキーマ→ドメインモデル」の順にすると、ドメイン設計の結果によってDBスキーマを作り直す（マイグレーションのやり直し）が発生し得るため避ける

### ステップ 5: ADR相当の判断を洗い出す

`detailed-design` スキルのステップ2.5と同様に、既存ADR（`docs/adr/`）でカバーされていない技術判断がないか確認する。

- 判断が必要な項目が見つかったら、**その判断が最初に確定するIssue**（多くはドメイン層のIssue）の設計メモに「実装前にADRを作成してから進める」と明記する
- その判断に依存する後続Issue（多くはInfrastructure層のIssue）には「先行Issueで決定したADRに従う」と明記する
- ADRの判断主体を後続Issue側に置かない（先に確定させる側が決定すべき判断を、後から追従する側に委ねると手戻りが起きる）

### ステップ 6: 分割案をユーザーに提示し、承認を得る

以下の形式で提示し、**承認を得るまでIssueを起票しない**（複数Issueの起票は取り消しコストが高いため、単一Issue作成時のような無確認起票はしない）。

```markdown
| # | タイトル(案) | レイヤー | 依存 |
|---|---|---|---|
| 1 | ... | Domain | なし |
| 2 | ... | Infrastructure | #1 |
```

ユーザーから粒度・順序の指摘があれば反映し、合意できるまで繰り返す。

### ステップ 7: 本文を組み立てて起票する

各sub-issueの本文は `create-issue` スキルの命名規則・テンプレート運用（`.github/ISSUE_TEMPLATE/feature.md` を読み込み `--body` で渡す、`--template`・`--label` は使わない）に従う。

加えて「設計メモ」に以下を必ず含める。

- 関連Issue: #親Issue番号（親Issue）
- 依存Issue: #NN（先行Issue）が完了していること（先行Issueがない場合は省略）
- 後続Issue: #NN（後続Issue）はこの完了後に進める（後続Issueがない場合は省略）

### ステップ 8: 依存順にIssueを作成し、前方参照を実番号に更新する

- 依存関係の順（先行Issueから）に `gh issue create` する
- 後続Issueの番号はまだ存在しないため、作成時点では本文中の「後続Issue」欄を仮の記述にしておく
- 全Issue作成後、`gh issue edit <番号> --body "..."` で「後続Issue」欄を実際の番号に更新する

---

## チェックリスト

- [ ] 分割案をユーザーに提示し承認を得たか
- [ ] 各sub-issueに「関連Issue: #親Issue（親Issue）」があるか
- [ ] 依存関係の順序がDependency Rule（Domain設計が先、Infrastructure実装が後）と整合しているか
- [ ] ADR相当の判断がある場合、それを最初に確定するIssueの設計メモに明記したか
- [ ] 前方参照（後続Issue）を実番号に更新したか
- [ ] タイトルが英語・本文が日本語になっているか（`create-issue` スキルのルールに従う）

## 参照ドキュメント

- `create-issue` スキル — 個別Issueの起票規約
- `detailed-design` スキル（ステップ2.5） — ADR要否の判断基準
- `docs/adr/ADR-001-clean-architecture.md` — Dependency Rule
- `docs/guidelines/development-flow.md` — PR分割ガイドライン（関心事の単位）
