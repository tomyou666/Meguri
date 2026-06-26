# Meguri — リブランド方針メモ

`scraperbot` / `scraper-bot` から **Meguri** へ改称するためのコンセプトと、後続 AI タスク（コードリネーム・アイコン生成など）用の指示書。

## ブランドコンセプト

### 一言

**Web を巡り、取り込み、つなげるデスクトップワークスペース。**

### 語源・意味

| 項目 | 内容 |
|------|------|
| 名称 | **Meguri**（巡り） |
| 読み | めぐり |
| 英字表記 | Meguri（先頭大文字は UI・タイトルバー用。識別子は小文字 `meguri`） |
| 核となるイメージ | リンクを辿る巡回、サイト内を回る、取得した断片が一つの流れにつながる |

### 製品の実態（名前が伝えるべきこと）

- Web ページとリンク先 PDF を取得し、Markdown 等に変換する
- グラフ上でノード（URL）を管理し、クロール・差分・エクスポートする
- Go バックエンド（CLI + ライブラリ）+ Wails3 デスクトップ UI のモノレポ
- ローカルファースト（SQLite、プロジェクトファイルのエクスポート/インポート）

### トーン

- **sharp** — 短く、覚えやすく、装飾過多にしない
- **technical** — 比喩は「巡る」程度に留め、機能を誤解させない
- 避ける語感: `bot`（チャットボットと混同）、`scraper` の直球（旧名の焼き直し感）

### ターゲット像

- サイトの構造を把握したい開発者・編集者・リサーチャー
- 単発スクレイプより「ワークスペースで巡回・蓄積・比較」するユーザー

---

## 命名規約（確定案）

| 用途 | 旧 | 新 |
|------|----|----|
| デスクトップアプリ表示名 | scraperbot | **Meguri** |
| Wails `Name` / ウィンドウタイトル | scraperbot | **meguri** または **Meguri**（UI 一貫で決める） |
| CLI バイナリ | `scraperbot` | **`meguri-cli`** |
| GitHub リポジトリ | `scraper-bot` | **`meguri`**（要: remote リネーム or 新 repo） |
| バックエンド Go module | `scraperbot` | **`meguri`** |
| フロント Go module | `scraperbot-front` | **`meguri`** または **`meguri-app`**（モノレポなら単一 `meguri` + サブパスも可） |
| CLI エントリ | `backend/cmd/scraperbot/` | **`backend/cmd/meguri-cli/`** |
| User-Agent 等 | `scraperbot/0.1` | **`meguri/0.1`**（バージョンは実装時の値に合わせる） |
| ローカル DB ファイル | `scraperbot.db` | **`meguri.db`** |
| XDG データディレクトリ | `~/.local/share/scraperbot` 等 | **`meguri`** |
| プロジェクトファイル拡張子 | `.scrb` | **`.crawl`** または **`.crawlproj`**（下記） |
| manifest `app` フィールド | `scraper-bot` | **`meguri`** |
| サービス名（Wails） | ScraperService 等 | 実装時に `CrawlService` 等へ段階リネーム可（必須ではない） |

### プロジェクトファイル拡張子

Grill では **機能ベースの英語拡張子** を選択。

| 候補 | 利点 |
|------|------|
| **`.crawlproj`**（推奨） | ZIP プロジェクトであることが明確。他ツールの `.proj` と同系 |
| `.crawl` | 最短。ただし単一ファイル形式と誤解されやすい |

**互換**: 移行期間は `.scrb` の読み込みを残す。エクスポートのデフォルトは新拡張子。

フォーマット仕様の正: [`docs/formats/scrb-v1.md`](formats/scrb-v1.md) → リネーム後は `crawlproj-v1.md` 等に改題。

---

## ビジュアル・アイコン方針

### 方向性

- **巡る・つながる・グラフ** のいずれか 1 つを主モチーフに（複数盛り込まない）
- デスクトップアプリアイコンとして 16〜1024px でシルエットが読めること
- ダーク UI 前提（現ウィンドウ背景 `rgb(27,38,54)` 付近）でもタスクバーで浮くコントラスト
- フラット過ぎず、軽い奥行きは可。絵文字・文字「M」依存は避ける（小サイズで潰れる）

### モチーフ案（いずれか 1 つを採用）

1. **環（わ）** — 巡回ループ。シンプルな円弧＋矢印
2. **グラフノード** — 3 点と細いエッジ。ワークスペースのグラフ UI と一致
3. **道筋** — 曲線パス上を進む点。リンク辿りのメタファー

### カラー（参考）

| 役割 | 値 |
|------|-----|
| 背景 | 深い青灰 `#1b2636` 付近（現 UI と調和） |
| アクセント | テール系 `#2dd4bf` / 青緑、または琥珀 `#f59e0b`（暖色で「巡り」の動き） |
| 線・シンボル | 高コントラストの off-white `#e8eaed` |

現状アイコンは Wails デフォルト（`front/build/appicon.png` 等）。差し替え対象:

- `front/build/appicon.png`
- `front/build/windows/icon.ico`
- `front/build/darwin/icons.icns`
- `front/build/ios/icon.png`
- `front/build/appicon.icon/`（macOS icon composer 用）

---

## アイコン生成プロンプト（コピペ用）

GPT Image 2 / 他の画像モデル向け。サイズはアイコン原稿なら `1024_1024`。

### A. 環＋巡回（推奨）

```
App icon for "Meguri", a desktop web-crawl workspace tool.
Minimal flat vector logo: a single smooth circular arc (almost a ring) with a small arrow
hinting continuous traversal; one accent node dot on the arc.
Dark blue-gray rounded-square background (#1b2636), teal accent (#2dd4bf),
off-white symbol (#e8eaed). No text, no letters, no emoji.
Centered, generous padding, readable at 32px. macOS/iOS app icon style, subtle soft shadow.
```

### B. グラフノード

```
Minimal app icon: three small circles connected by thin lines forming a loose triangle path,
suggesting a crawl graph. Dark blue-gray background, teal nodes, light gray edges.
Flat design with slight depth, no typography, square format with rounded corners,
professional developer tool aesthetic.
```

### C. 道筋・リンク辿り

```
Abstract app icon: a luminous dot traveling along a gentle S-curve path on dark background,
metaphor for following links across the web. Clean geometric style, teal and white on #1b2636,
no text, high contrast silhouette, suitable for Windows and macOS dock icons.
```

### 編集用（ベース画像がある場合）

```
Keep the overall composition and palette unchanged; refine edges for small-size clarity,
increase contrast between symbol and background, ensure the symbol reads clearly at 32x32 pixels.
```

---

## コードリネーム — AI 向けタスクプロンプト

後で別セッションに渡す用。実施前にバックアップ・ブランチ作成を推奨。

```
プロジェクト scraper-bot を Meguri にリブランドする。方針は docs/meguri-rebrand.md に従う。

## ゴール
- 表示名・ドキュメント・Go module パス・バイナリ名・識別子を meguri 系に統一
- デスクトップアプリが主、CLI は meguri-cli
- .scrb → .crawlproj（.scrb 読み込み互換は維持）

## 主な変更箇所
- README.md, backend/README.md, front/README.md
- backend/go.mod (module scraperbot → meguri)
- front/go.mod (module scraperbot-front → meguri 等)
- backend/cmd/scraperbot/ → cmd/meguri-cli/
- front/main.go (Wails Name, Title, Description)
- front/internal/app/config.go (dbFileName)
- front/internal/app/dbpath_prod.go (XDG ディレクトリ名)
- front/internal/infrastructure/scrb/ → crawlproj パッケージ名も検討
- front/shared/defaults.json (User-Agent)
- front/frontend/src/i18n/messages.ts (UI 文言)
- docs/formats/scrb-v1.md, docs/api/scraper-ui.md
- .vscode/launch.json, tasks.json（デバッグ構成名・パス）
- Wails bindings 再生成 (make -C front bindings)

## 制約
- 動作を壊さない。リネーム後に backend テスト・front テストを実行
- .scrb インポートは当面残す
- 既存ユーザーの scraperbot.db は初回起動時に meguri.db へ移行するか、読み込みパスを両方見る（要判断・実装時にコメント）
- git commit はユーザーが明示したときだけ

## 検索キーワード（grep 用）
scraperbot, scraper-bot, ScraperBot, scraper_bot, .scrb, scraperbot-front, ScraperService
```

---

## リネーム影響チェックリスト

実施時に順に確認。

- [ ] ルート・各 README のタイトルと説明
- [ ] `go.mod` / import パス一括（`backend/`, `front/`, `front/tools/`）
- [ ] `cmd/scraperbot` ディレクトリ名と Makefile / ドキュメント内のビルド手順
- [ ] Wails アプリ名・ウィンドウタイトル・`Description`
- [ ] SQLite ファイル名・マイグレーションパス
- [ ] `.scrb` 拡張子フィルタ（ダイアログ・エクスポート）と `manifest.app`
- [ ] `front/frontend/bindings/` 再生成
- [ ] i18n メッセージ（`messages.ts`）
- [ ] VS Code launch / task 名
- [ ] `backend/doc/` 設計書内の旧名称（必要なら）
- [ ] GitHub remote URL（repo リネーム後）
- [ ] アイコン資産差し替え（`front/build/`）

---

## Grill で確定した意思決定（経緯）

| 質問 | 回答 |
|------|------|
| 名称が伝えるもの | technical（機能を正確に） |
| 言語 | 日本語ローマ字（カタカナ語） |
| 核となる動作 | kuroru 系 → **meguri（巡り）に最終決定** |
| 接尾辞 | なし（メインブランド） |
| 変更範囲 | フル（コード・repo・拡張子まで） |
| CLI と GUI | GUI 主、CLI は `meguri-cli` |
| 拡張子 | 機能ベース英語（`.crawl` / `.crawlproj`） |
| トーン | sharp |
| web 前置 | なし |
| 綴り | カタカナ語ローマ字 **meguri** |

---

## 参考: 現状の主要パス

```
scraper-bot/                 # リポジトリルート（→ meguri）
├── backend/
│   ├── go.mod               # module scraperbot
│   └── cmd/scraperbot/      # CLI エントリ
├── front/
│   ├── go.mod               # module scraperbot-front
│   ├── main.go              # Wails Name: scraperbot
│   └── build/               # アイコン（要差し替え）
└── docs/formats/scrb-v1.md  # .scrb 仕様
```
