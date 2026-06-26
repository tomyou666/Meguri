# DESIGN.md → shadcn CSS 参照

## 標準マッピング（shadcn 組み込み変数）

| shadcn 変数 | DESIGN.md 候補 |
|-------------|----------------|
| `--background` | `canvas` |
| `--foreground` | `on-dark` / `ink` |
| `--card` | `surface-card` |
| `--card-foreground` | `on-dark` |
| `--popover` | `surface-elevated` |
| `--popover-foreground` | `on-dark` |
| `--primary` | `primary` |
| `--primary-foreground` | `on-primary` |
| `--secondary` | `surface-soft` |
| `--secondary-foreground` | `on-dark` |
| `--muted` | `carbon-gray` / `surface-elevated` |
| `--muted-foreground` | `muted` / `body` |
| `--accent` | `surface-elevated` |
| `--accent-foreground` | `on-dark` |
| `--destructive` | `m-red` |
| `--border` | `hairline` |
| `--input` | `hairline` |
| `--ring` | `bmw-blue` / `m-blue-dark` |

## セマンティック拡張（追加必須）

| 追加変数 | DESIGN.md ソース | 用途例 |
|----------|------------------|--------|
| `--success` | `success` | 成功バッジ、完了状態、toast.success |
| `--success-foreground` | `on-dark` または `on-primary` | success 上の文字 |
| `--warning` | `warning` | 警告バッジ、注意 toast |
| `--warning-foreground` | `on-dark` または `canvas` | warning 上の文字 |
| `--info` | `info`（無ければ `bmw-blue` / `electric-blue`） | 案内 Alert、toast.info |
| `--info-foreground` | `on-dark` | info 上の文字 |

### foreground の選び方

- **半透明背景**（`bg-success/20`）: foreground は `--success` 本体色を `text-success` に使う
- **solid 背景**（`bg-success`）: `--success-foreground` を `text-success-foreground` に使う
- コントラスト不足なら DESIGN の `on-dark` / `on-primary` で調整

## chart / sidebar（明示同期時のみ）

| 変数 | 候補 |
|------|------|
| `--chart-1` | `m-blue-light` |
| `--chart-2` | `m-blue-dark` / `bmw-blue` |
| `--chart-3` | `m-red` |
| `--chart-4` | `electric-blue` |
| `--chart-5` | `muted` / `carbon-gray` |
| `--sidebar-primary` | `bmw-blue` |

## oklch 変換

hex → oklch はブラウザ DevTools、または `colorjs.io` 等で変換。ターゲット CSS が既に oklch なら **必ず oklch に揃える**。

例（このリポジトリの DESIGN.md）:

| hex | 用途 |
|-----|------|
| `#0fa336` | success |
| `#f4b400` | warning |
| `#1c69d4` | info（bmw-blue） |

## Badge 置換例

```tsx
// Before
success: 'border-transparent bg-emerald-500/20 text-emerald-400',
warning: 'border-transparent bg-amber-500/20 text-amber-400',

// After
success: 'border-transparent bg-success/20 text-success',
warning: 'border-transparent bg-warning/20 text-warning',
```
