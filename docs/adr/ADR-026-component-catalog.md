# ADR-026: コンポーネントカタログツール

- ステータス: 承認済

## コンテキスト

packages/ui にUIコンポーネントを集約するMonorepo構成（ADR-002）を採用している。コンポーネントが増えるにつれて、以下の課題が生じる。

- 実装済みコンポーネントの一覧・バリアントを確認する手段がなく、重複実装が発生しやすい
- デザインと実装の乖離をレビュー時に発見しにくい
- コンポーネント単体の動作確認にアプリ全体の起動が必要

これを解消するため、コンポーネントカタログツールを導入する。技術スタックの前提は以下のとおり。

- ADR-005: React 19 + TypeScript
- ADR-006: Vite（apps/web、packages/ui ともにViteベース）
- ADR-015: @base-ui/react + class-variance-authority（shadcn/uiではなくBase UIプリミティブを直接採用）
- ADR-017: Biome + Oxlint（コード品質ツール）

## スコープ外

- ビジュアルリグレッションテストツール（Chromatic 等）の導入判断は本ADRの対象外
- アクセシビリティ自動検査ツールの選定は本ADRの対象外
- デザイントークンの管理方針は本ADRの対象外

## 決定

コンポーネントカタログツールとして **Storybook**（`@storybook/react-vite` ビルダー）を採用する。

## 検討した選択肢

### 選択肢1: Storybook

#### 概要

コンポーネントカタログのデファクトスタンダード。CSF（Component Story Format）でStoryを記述し、`@storybook/react-vite` ビルダーによりViteでビルドする。Storybook 8以降でReact 19をサポート。

#### メリット

- React 19・Vite 8・Tailwind CSS 4 のいずれにも対応実績があり、互換性リスクが低い
- `@storybook/react-vite` により既存のVite設定（Tailwind CSS PostCSS等）をそのまま再利用できる
- Monorepoでのpackages/ui配置に対するリファレンス実装が豊富で、セットアップコストが低い
- アドオン（`@storybook/addon-a11y`、`@storybook/addon-interactions` 等）が充実しており、後からアクセシビリティ検査・インタラクションテストを追加できる
- CSF形式はStorybookに固有ではなく、Ladle等の他ツールでも読める可換なフォーマット

#### デメリット

- アドオンが多い分、起動が遅くなりやすい（Viteビルダーでの開発サーバーは十分高速だが、webpack時代より引き継がれたオーバーヘッドが残る）
- 設定ファイル（`.storybook/`）が多く、初期セットアップの学習コストがある

### 選択肢2: Ladle

#### 概要

Viteネイティブに設計されたStorybookの軽量代替。Storybookと同じCSF形式のStoryをそのまま読み込める。設定ファイルが最小限で、起動が高速。

#### メリット

- Viteネイティブ設計のため起動・HMRが最速
- 設定が少なく、初期セットアップが容易
- CSF形式互換のためStorybookへの移行コストが低い

#### デメリット

- アドオンエコシステムがStorybookに比べて大幅に小さく、アクセシビリティ検査・インタラクションテスト等を後から追加しにくい
- React 19 + Vite 8 + Tailwind CSS 4 という最新スタック組み合わせでの採用事例が少なく、互換性リスクがある
- メンテナンスがStorybookに比べて限定的で、長期的な安定性に不安がある

### 選択肢3: Histoire

#### 概要

Vue向けに設計されたコンポーネントカタログ。Reactサポートは実験的フェーズ（v0.x）にあり、正式サポートには至っていない。

#### メリット

- Vue SFCとの統合が優れており、Vue中心プロジェクトでは最良の選択肢

#### デメリット

- ReactサポートがExperimentalであり、本プロジェクト（ADR-005: React採用）では採用リスクが高い
- Tailwind CSS 4・React 19 との組み合わせ事例がほぼなく、動作保証が取れない

## 決定理由

- **互換性の確実性**: React 19 + Vite 8 + Tailwind CSS 4 という最新スタックにおいて、Storybookは `@storybook/react-vite` ビルダーを通じた対応実績が最も豊富。Ladleは軽量だが、この組み合わせでの事例が少なく互換性リスクがある
- **既存Vite設定の再利用**: `@storybook/react-vite` はプロジェクトのVite設定（Tailwind CSS PostCSS含む）を継承するため、二重管理が発生しない
- **Monorepo対応**: packages/ui を対象としたStorybook構成のリファレンスが豊富で、ワークスペースパッケージの解決も標準的に動作する
- **CSF形式の可換性**: StorybookのStoryはCSF形式で記述するため、将来ツールを変更しても移行コストが限定的
- **拡張性**: 現時点ではStory作成が主目的だが、アクセシビリティ検査（`@storybook/addon-a11y`）やインタラクションテストを後から追加できるアドオン基盤があることは長期的に有益

Ladleは起動速度の面では有利だが、アドオンエコシステムの不足と最新スタックでの互換性リスクからStorybookに劣ると判断する。HistoireはReactサポートが実験的であり選択肢から除外する。

## 結果

### 良い影響

- packages/ui の全コンポーネントとそのバリアントをカタログで一覧できるため、重複実装を防ぎ再利用を促進できる
- コンポーネント単体をアプリ全体の起動なしに確認・レビューできるため、開発・レビュー効率が向上する
- `@storybook/react-vite` により既存のVite設定を再利用でき、Tailwind CSSのスタイルがカタログ上でも正しく反映される
- 将来的にアクセシビリティ検査・インタラクションテストをアドオンで追加できる

### 悪い影響

- `.storybook/` 設定ファイルの管理が増え、Storybook自体のアップデート追従コストが発生する
- Storyファイル（`*.stories.tsx`）の作成・維持がコンポーネント追加時の作業に加わる
