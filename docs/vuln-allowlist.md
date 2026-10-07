# 脆弱性許容リスト（vuln-gate）

`make vuln` は govulncheck と npm audit の生の検出結果をそのまま終了コードにします。CI とゲート付きのローカル確認は `make vuln-gate` を使います。ゲートはスキャナーを実行し、検出をリポジトリルートの [`vuln-allowlist.yaml`](../vuln-allowlist.yaml) と突合します。

## 使い分け

| コマンド | 用途 |
| --- | --- |
| `make vuln` | 許容なしの生検査。検出があれば失敗 |
| `make vuln-gate` | 許容リストと突合。CI（[`.github/workflows/quality.yml`](../.github/workflows/quality.yml)）の成否 |

## スキャナー範囲

- npm: `front/frontend` の `npm audit --json`。重大度 **high** と **critical** のみ判定
- Go: `backend`、`front`、`tools/*/go.mod` 各ディレクトリで `govulncheck -json ./...`

バイナリ名は環境変数 `NPM` / `GOVULNCHECK` で上書きできます（Makefile と同名）。

## 検出の単位

検出 1 件は **advisory 1 件**（ecosystem + id + package + module）です。npm audit が同じ GHSA を依存先パッケージごとに数えても、それらは経路であり別件にはしません。

## 許容エントリの必須項目

| 項目 | 意味 |
| --- | --- |
| `id` | advisory ID（npm は `GHSA-...`、Go は `GO-...`） |
| `ecosystem` | `npm` または `go` |
| `package` | advisory のパッケージ名（npm）または脆弱なモジュールパス（Go） |
| `module` | リポジトリルートからのパス（例: `front/frontend`、`backend`、`tools/url-fetch`） |
| `via` | その advisory に到達する **直接依存の集合**（順序不問）。モジュール自身の問題なら `[]` |
| `reason` | 許容する理由 |
| `expires` | `YYYY-MM-DD`。**Asia/Tokyo** の暦日でその日を含む。翌日以降は失敗 |

同じ `id` + `ecosystem` + `package` + `module` の重複は設定エラーです。

### `via` の求め方

- npm: `package.json` の `dependencies` / `devDependencies` / `optionalDependencies` のうち、audit の effects 木を上へ辿って到達する名前の集合
- Go: `go.mod` の indirect ではない require のうち、finding の trace に現れるモジュールの集合。脆弱なモジュールが対象モジュール自身なら `via: []`

リストの `via` と検出の `via` が完全一致しないと失敗します。新しい直接依存からも届くようになった場合は、エントリを更新するか修正します。

## ゲートが失敗する条件

- 許容リストに無い検出
- `via` の不一致
- `expires` の翌日以降（Asia/Tokyo）
- 検出に対応しない（消えた）許容エントリ
- スキャナー起動失敗、または JSON を返せない

## エントリの追加手順

1. `make vuln` または `make vuln-gate` の出力で id / package / module / via を確認する
2. [`vuln-allowlist.yaml`](../vuln-allowlist.yaml) に理由と期限付きで追加する
3. `make vuln-gate` が成功することを確認する
4. 修正版が出たら依存を更新し、不要になったエントリを削除する（残すと stale で失敗する）

## 実装

ゲート本体は [`tools/vuln-gate`](../tools/vuln-gate) です。
