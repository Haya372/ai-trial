# ADR-005: フロントエンドフレームワーク

- ステータス: 承認済

## コンテキスト

本プロジェクトはPC/タブレットブラウザ向けのカレンダーアプリであり、以下のUI要件を満たすフロントエンドフレームワークが必要となる。

- 月ビュー・週ビュー・タイムライン（PRD-002, PRD-007）といった複雑なグリッド/時間軸UIを構築する
- ビュー切り替えが1秒以内（PRD-002 制約）で完了するレンダリング性能
- クイック登録（PRD-005）のようなインタラクティブな操作
- ブラウザ通知（PRD-006 スマートリマインダー）でNotification APIを扱う
- 予定共有URL（PRD-004）のような公開ページも配信する（ただし本プロジェクトの中心はログイン後の対話UIでありSEO要件は薄い）
- PC/タブレット最適化要件でありスマホ非対応
- ADR-001（Clean Architecture）のAdapter層以外にフレームワーク依存を漏らさない
- ADR-002（Monorepo）で他パッケージと共存しやすい

候補としてReact + TypeScript、Vue.js + TypeScript、Next.js（React SSR）を検討する。

## スコープ外

- フロントエンドビルドツールの選定はADR-006で扱う
- フロントエンド状態管理ライブラリの選定はADR-007で扱う
- コンポーネント設計・デザインシステム（Tailwind、Chakra UI等）の選定は詳細設計時に決定する
- カレンダー描画に外部ライブラリ（FullCalendar等）を利用するかどうかの判断は詳細設計時に決定する
- E2Eテストフレームワーク（Playwright、Cypress等）の選定は品質保証設計時に別途決定する
- ホスティング環境・i18n・a11yライブラリの選定は本ADRの対象外

## 決定

フロントエンドフレームワークとして **React 18以降 + TypeScript** を採用する。SPA（Single Page Application）としてバックエンド（ADR-003のGo）と分離してデプロイする。

## 検討した選択肢

### 選択肢1: React + TypeScript

#### 概要

Metaが開発する宣言的UIライブラリ。JSX + TypeScriptで宣言的にコンポーネントを記述し、SPAとしてバックエンドと分離してデプロイする構成が一般的。React 18からconcurrent renderingが導入されている。

#### メリット

- カレンダー関連ライブラリのエコシステムが最も豊富（FullCalendar、react-big-calendar、@nivo/calendar、date-fns等）で、複雑なカレンダーUI（PRD-002, PRD-007）の構築コストを抑えられる
- TypeScript対応が最も成熟しており、コンポーネントのProps・状態の型付けが確実
- React 18のconcurrent renderingにより、月ビュー→週ビューの切り替え（PRD-002 制約: 1秒以内）で大きなDOM更新をユーザー操作を止めずに処理できる
- Notification API（PRD-006）はブラウザ標準APIのため、Reactからも直接扱える
- Monorepoツール（Turborepo、Nx等）でReactはリファレンス実装が豊富で、ADR-002の運用と相性が良い
- Adapter層としてReactを閉じ込めやすく、UseCase呼び出しをカスタムフックに集約することでClean Architecture（ADR-001）との親和性を保てる

#### デメリット

- ReactはUIライブラリでありフレームワーク的な規約が少ないため、状態管理・ルーティング・データフェッチを個別に選定する必要がある（ADR-006/007で別途決定）
- 規約が少ない分、チーム内での実装スタイルが揺れやすく規約整備が必要
- Vueに比べてテンプレート構文がなく、JSX + TypeScriptの学習コストがやや高い

### 選択肢2: Vue.js 3 + TypeScript

#### 概要

Evan Youが開発する漸進的採用可能なUIフレームワーク。SFC（Single File Component）による構造化とComposition API + TypeScriptによる型安全な開発が可能。

#### メリット

- SFC（テンプレート + スクリプト + スタイル）により、コンポーネントの責務が視覚的に明確
- Vue公式のツール（Vue Router、Pinia）が揃っており、意思決定が少なく済む
- Composition APIとリアクティビティシステムにより、カレンダーUIの状態管理が直感的
- 学習コストがReactに比べて低い

#### デメリット

- カレンダー関連ライブラリのエコシステムがReactに比べて狭い（FullCalendarはVueバインディングがあるが、選択肢が少ない）
- TypeScript対応は成熟してきたものの、Reactに比べてテンプレート内の型推論の精度でまだ差がある
- 大規模カレンダーUIのリファレンス実装がReactに比べて少なく、パフォーマンスチューニングの情報が限定的
- 状態管理ライブラリ（TanStack Query等）のVue版はReact版に比べて機能追従が遅れる傾向がある
- Monorepoツールでのリファレンスや事例がReactより少ない

### 選択肢3: Next.js（React SSR/RSC）

#### 概要

VercelがメンテナンスするReactメタフレームワーク。SSR（Server Side Rendering）・SSG（Static Site Generation）・RSC（React Server Components）を標準サポートし、App Routerによるファイルベースルーティングを提供する。

#### メリット

- 予定共有URL（PRD-004）などの公開ページでSSR/SSGを活用でき、初回表示が速い
- ファイルベースルーティング・データフェッチ・画像最適化などの機能が揃っており、標準的な構成が決まりやすい
- Reactエコシステム（カレンダーライブラリ、状態管理）をそのまま利用できる
- App Router + Server Componentsにより、データフェッチをサーバー側に寄せられる

#### デメリット

- 本プロジェクトはPC/タブレットブラウザ向けの認証必須アプリ（PRD-001）であり、ログイン後の画面ではSSRの恩恵が薄い
- Next.js自身がNode.jsサーバーを必要とし、バックエンド（ADR-003のGo）とは別のランタイムを運用することになる（フロントエンドの配信がSPAより複雑になる）
- ADR-003で決定したGoバックエンドとNext.js（Node.js）の二重運用となり、Monorepo内で2つのサーバーサイドランタイムを管理する必要が生じる
- Clean ArchitectureのAdapter層としてNext.jsを閉じ込めようとすると、App Router固有の規約（Server Components、Route Handlers）が層をまたぐ設計上の複雑さを生む
- 予定共有URL程度のSSR要件であれば、SPA + バックエンドAPIで十分対応可能

## 決定理由

- **SSR不要の要件**: 本プロジェクトの中心機能（PRD-001〜PRD-009）はログイン後の対話的UIであり、公開ページは予定共有URL（PRD-004）に限られる。共有URLはバックエンド（ADR-003のGo）が最小限のOG情報付きHTMLを返す設計で十分対応可能で、Next.jsのSSR/RSC機能を導入する必要性が低い
- **カレンダーライブラリのエコシステム**: 月ビュー・週ビュー・タイムライン（PRD-002, PRD-007）といった複雑なカレンダーUIの構築において、React系のカレンダーライブラリ（FullCalendar、react-big-calendar等）が最も豊富。Vueは事例・選択肢ともにReactに劣る
- **パフォーマンス要件**: ビュー切り替え1秒以内（PRD-002）を安定達成するには、React 18のconcurrent renderingが有効。大きなDOM更新をユーザー操作を止めずに処理できる
- **Clean Architectureとの整合**: UseCase呼び出しをカスタムフックに集約し、コンポーネントからは「フックを呼ぶだけ」の形にすることで、React依存をAdapter層（フック層）に閉じ込められる。ADR-001の狙いと一致する
- **Monorepo適合性**: React（SPA）の構成はMonorepoツールのリファレンス実装が最も多く、ADR-002の運用と相性が良い
- **TypeScript対応の成熟度**: Reactは型定義の充実・型推論の精度・エディタ支援のいずれも最も成熟しており、Domain層と対応するViewModelの型安全性を担保しやすい

Vueは開発者体験の良さはあるが、カレンダーライブラリのエコシステムでReactに劣ることから見送る。Next.jsは要件に対して機能過多であり、Goバックエンドとの二重ランタイム運用の複雑さから見送る。

## 結果

### 良い影響

- Reactエコシステムのカレンダーライブラリ・UIコンポーネントを幅広く選定でき、複雑なカレンダーUIの構築コストを抑えられる
- SPAとして構築することでフロントエンド配信が単純化し、Goバックエンド（ADR-003）とランタイム分離が明確になる
- カスタムフックにUseCase呼び出しを集約する設計により、Reactへの依存をAdapter層に閉じ込められる（ADR-001と整合）
- React 18のconcurrent renderingによりビュー切り替え時の体感速度が確保しやすい

### 悪い影響

- SPAのため初回ロードで全JSをダウンロードする必要があり、初回表示速度はSSRに劣る（本プロジェクトの用途では許容範囲）
- ルーティング・状態管理・データフェッチを個別に選定する必要があり、Next.js/Vueに比べて初期セットアップの意思決定コストが高い（ADR-006, ADR-007で別途決定する）
- 予定共有URL（PRD-004）のOG情報付きHTMLは、バックエンド側で最小限のテンプレート配信を別途設計する必要がある
- ReactはUIライブラリでありフレームワーク規約が少ないため、実装スタイルの一貫性はチームで規約整備が必要
