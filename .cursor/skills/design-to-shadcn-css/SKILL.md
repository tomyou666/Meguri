---
name: design-to-shadcn-css
description: Reads a specified DESIGN.md and syncs its color tokens into shadcn/ui CSS variables (:root and .dark) in globals.css or index.css, including semantic extensions success, warning, and info. Use when applying DESIGN.md to shadcn theme variables, syncing design tokens to CSS, or mapping semantic colors (background, foreground, primary, destructive, border, input, success, warning, info).
disable-model-invocation: true
---

# DESIGN.md → shadcn/ui CSS 変数

指定されたDESIGN.md の内容を読み取り、指定されたglobals.css（または index.css）にある shadcn/ui のCSS変数定義（:root と .dark）に正確に反映してください。shadcn/uiが要求するセマンティックカラー（background, foreground, primary, destructive, border, input等）に適切にマッピングして上書きしてください。

あわせて **success / warning / info** を CSS 変数として追加し、`@theme inline` に Tailwind ユーティリティ（`bg-success` 等）を配線してください。コンポーネント内の `emerald-*` / `amber-*` などパレット直書きは、トークン追加後に置き換える。

## 入力の特定

ユーザーがパスを省略した場合は確認する。両方指定されている場合はそのパスを使う。

| 入力 | 例 |
|------|-----|
| デザイントークン | リポジトリルートの `DESIGN.md`、またはユーザー指定パス |
| CSS ターゲット | `src/globals.css`、`src/index.css` などユーザー指定パス |

## ワークフロー

1. **DESIGN.md を読む**
   - 先頭 YAML の `colors:`（および必要なら `rounded:`）を抽出する
   - 本文の `## Colors` などでセマンティック用途（CTA・hairline・dark-only 等）を補足として参照する

2. **ターゲット CSS を読む**
   - `:root` と `.dark` ブロック内の `--*` 変数を更新対象とする
   - `@theme inline` は success / warning / info の `--color-*` 追加時のみ追記（それ以外は維持）
   - `@layer base`・`@import` は触らない

3. **マッピングして上書き**
   - ターゲットファイルが `oklch(...)` 形式なら、hex を **oklch に変換**して既存形式に揃える（変換不能な場合のみ hex を使う）
   - **success / warning / info** を追加（下記「セマンティック拡張」）

4. **検証**
   - `:root` / `.dark` の両方に、shadcn が参照する主要変数が欠けていないか確認する
   - DESIGN が dark-only と明記している場合は「ダークモード方針」に従う（下記）

## ダークモード方針

DESIGN.md に light 用パレットが無く「near-black canvas」「no light-mode」とある場合:

- **推奨**: `:root` と `.dark` の両方に同じダークパレットを適用する（アプリが常にダーク UI のとき）
- ユーザーがライト/ダーク切替を求める場合のみ、`:root` をライト用に分離する（そのときは DESIGN から推論できる範囲で対比を作る）

## セマンティック拡張（success / warning / info）

shadcn 標準には含まれないが、DESIGN.md の semantic 色と UI 実装（Badge・toast・Alert 等）のため **必ず追加**する。

### 1. DESIGN.md からのマッピング

| CSS 変数 | DESIGN.md ソース | 無いときの推論 |
|----------|------------------|----------------|
| `--success` | `colors.success` | 本文 `## Colors` の Semantic / Success |
| `--warning` | `colors.warning` | 本文の Semantic / Warning |
| `--info` | `colors.info`（あれば） | `bmw-blue` または `electric-blue`（情報・リンク系アクセント） |

各トークンに **foreground** ペアも定義する（`--success-foreground` 等）。暗背景 UI では本文色を `on-dark` / `primary`（白）に、明るい fill ボタンでは `on-primary`（黒）に合わせる。

### 2. `:root` / `.dark` への追加

```css
--success: oklch(...);
--success-foreground: oklch(...);
--warning: oklch(...);
--warning-foreground: oklch(...);
--info: oklch(...);
--info-foreground: oklch(...);
```

- `:root` と `.dark` **両方**に同じ値を入れる（dark-only DESIGN のとき）
- 既に `--destructive` がある場合と同様、**塗り色本体**を `--success` 等に置く（`/20` の半透明はコンポーネント側で `color-mix` や opacity ユーティリティを使う）

### 3. `@theme inline` への配線

既存の `--color-primary: var(--primary)` と同パターンで追加:

```css
--color-success: var(--success);
--color-success-foreground: var(--success-foreground);
--color-warning: var(--warning);
--color-warning-foreground: var(--warning-foreground);
--color-info: var(--info);
--color-info-foreground: var(--info-foreground);
```

これで `bg-success/20`、`text-warning`、`border-info` 等が Tailwind v4 で使える。

### 4. コンポーネント追随（ユーザーが UI 同期を求める場合）

- `badge.tsx` の `success` / `warning` variant → `bg-success/20 text-success` 等に変更
- `emerald-*` / `amber-*` / `sky-*` 直書きをトークン参照に置換
- `info` は Alert・toast（sonner）・インライン案内に使う

### 5. 禁止事項

- DESIGN に無い色相（Tailwind `emerald` 等）を新規導入しない
- success / warning / info を **primary CTA の塗り**に使わない（DESIGN のボタン規約と同様、状態表示・バッジ・通知向け）
- M トライカラーは success / warning / info の代替にしない

## 編集ルール

- **上書きのみ**: 既存の変数名・ブロック構造・宣言順は可能な限り維持する
- **触らない**: `--font-*`、`@custom-variant`、chart/sidebar / success-warning-info 以外の `@theme` キー
- **@theme に追加可**: `--color-success*` / `--color-warning*` / `--color-info*`（セマンティック拡張時）
- **radius**: DESIGN の `rounded:` があるときのみ `--radius` を更新（例: `md: 6px` → `0.375rem`）
- **M トライカラー**（`m-blue-light` 等）は CTA/背景に使わない — `chart-*` や装飾用 CSS 変数（プロジェクトで定義済みの場合）に限定

## 完了チェックリスト

- [ ] DESIGN.md の `colors:` トークンを漏れなく参照した
- [ ] `:root` と `.dark` の `--background` 〜 `--sidebar-ring`（存在するもの）を更新した
- [ ] `--success` / `--warning` / `--info` と各 `-foreground` を追加した
- [ ] `@theme inline` に `--color-success*` / `--color-warning*` / `--color-info*` を配線した
- [ ] `primary` / `destructive` / `border` / `input` / `muted-foreground` が用途と矛盾しない
- [ ] 色形式がターゲット CSS の既存表記（oklch 等）と一致している
- [ ] Tailwind パレット直書き（`emerald-*` 等）をトークンに置換した（UI 同期時）

## 追加リソース

- トークン対応表・oklch 変換・chart/sidebar の割当: [reference.md](reference.md)
