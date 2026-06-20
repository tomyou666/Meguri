# scraperbot-front

Wails v3 デスクトップ UI。Go バックエンドはレイヤード構成（`internal/domain` + `infrastructure` + `usecase/wails_service`）、永続化は SQLite（GORM + GORM Gen + golang-migrate）。

## 構成

| パス | 役割 |
|------|------|
| `main.go` | エントリポイント |
| `internal/app/` | Wire composition root、dbpath、migrate |
| `internal/domain/` | ビジネスロジック |
| `internal/infrastructure/` | リポジトリ、`.scrb`、dialog |
| `internal/usecase/wails_service/` | Wails 公開サービス（`StoreService` / `ProjectService`） |
| `internal/model/` | Wails DTO（手書き）+ GORM Gen 生成 struct |
| `internal/query/` | GORM Gen 生成クエリ |
| `storage/schema.sql` | DDL 参照（`migrations/` と内容を一致させる） |
| `frontend/` | React + TypeScript UI |

## 前提

- Go 1.26+（[`.prototools`](../.prototools) / [proto](https://moonrepo.dev/proto) 推奨）
- Node.js + npm（同上）
- 開発用 Go CLI: `make tools`（`go.mod` の `tool` で `dlv` / `migrate` / `wails3` を管理）

Go の SQLite / golang-migrate は **CGO 不要**の pure Go スタック（`glebarez/sqlite` + `modernc.org/sqlite`）。ビルドタグの指定は不要です。

## 初回セットアップ

```powershell
cd front
make setup          # npm install
make tools          # go mod download（dlv, migrate, wails3）
make migrate-up     # DB 作成 + スキーマ適用
make gen            # GORM Gen（マイグレーション後に実行）
make wire           # wire_gen.go（internal/app 実装後）
make bindings       # TS bindings（wails_service 実装後）
```

## マイグレーション

| 項目 | 内容 |
|------|------|
| SQL の正（運用） | `internal/app/migrations/*.sql`（`NNNNNN_description.up.sql` / `.down.sql`） |
| SQL の正（参照） | `storage/schema.sql` |
| 起動時 | `app.Initialize` 内で `RunMigrations` → `m.Up()`（毎回バージョン確認・未適用分のみ適用） |
| 手動適用 | `make migrate-up` |
| 1 段戻す | `make migrate-down` |
| バージョン確認 | `make migrate-version` |

### 生成・更新されるもの

| 出力 | パス / 場所 | git 管理 |
|------|-------------|----------|
| SQLite DB ファイル | `data/scraperbot.db`（dev） | 対象外（`.gitignore`） |
| マイグレーション履歴 | DB 内 `schema_migrations` テーブル | — |

本番ビルドでは `dbpath_prod` により OS のアプリデータ領域に DB を配置します。

### スキーマ変更手順

1. `internal/app/migrations/` に新しい `.up.sql` / `.down.sql` を追加
2. `make migrate-up`（またはアプリ起動で自動適用）
3. 必要なら `tools/gen/main.go` の `GenerateModel(...)` を更新
4. `make gen`
5. `storage/schema.sql` を追随更新
6. `internal/infrastructure/persistence` / `internal/domain` を更新

## Wails bindings（TypeScript）

Go の Wails 公開 API から、フロント用の型安全 TS モジュールを生成します。

| 項目 | 内容 |
|------|------|
| コマンド | `make bindings`（`wails3 generate bindings -ts`） |
| 実行タイミング | `usecase/wails_service` のメソッド追加・変更、`internal/model` の DTO 変更後 |

### 生成されるファイル（手編集禁止）

| パス | 内容 |
|------|------|
| `frontend/bindings/scraperbot-front/index.ts` | サービス export 集約 |
| `frontend/bindings/scraperbot-front/internal/usecase/wails_service/storeservice.ts` | `StoreService` RPC |
| `frontend/bindings/scraperbot-front/internal/usecase/wails_service/projectservice.ts` | `ProjectService` RPC |
| `frontend/bindings/scraperbot-front/internal/usecase/wails_service/index.ts` | 上記 re-export |
| `frontend/bindings/scraperbot-front/internal/model/models.ts` | Go DTO の TS 型 |
| `frontend/bindings/scraperbot-front/internal/model/index.ts` | model re-export |
| `frontend/bindings/github.com/wailsapp/wails/v3/internal/eventcreate.ts` | Wails イベント登録 |
| `frontend/bindings/github.com/wailsapp/wails/v3/internal/eventdata.d.ts` | イベント型 |
| `frontend/bindings/encoding/json/` | JSON 補助型（DTO が必要とする場合） |

`frontend/src/adapters/compositeScraperAdapter.ts` が上記 bindings を import します。

## スキーマ生成（GORM Gen）

マイグレーション済み DB を introspect し、型安全なモデルとクエリ API を生成します。参考: `tmp/db-sample`（ローカル、gitignore）。

| 項目 | 内容 |
|------|------|
| コマンド | `make gen`（`go run ./tools/gen`） |
| 前提 | `make migrate-up` 済みの `data/scraperbot.db` |
| 設定 | `OutPath: ./internal/query`, `ModelPkgPath: model` |

### 生成されるファイル（手編集禁止）

| パス | 内容 |
|------|------|
| `internal/model/*.gen.go` | テーブル行 struct |
| `internal/query/gen.go` | `Use(db)` エントリ |
| `internal/query/*.gen.go` | テーブル別型安全クエリ |

### 手書き（Gen と混在させない）

| パス | 内容 |
|------|------|
| `internal/model/api.go` 等 | Wails 公開 DTO（bindings の型元） |

## Wire 生成

| 項目 | 内容 |
|------|------|
| コマンド | `make wire` |
| 生成物 | `internal/app/wire_gen.go`（手編集禁止、コミット対象） |

`providers.go` / `wire.go` 変更後に再実行してください。

## 生成物一覧

| 種別 | コマンド | 出力先 | 手編集 |
|------|----------|--------|--------|
| DB スキーマ | `make migrate-up` / 起動時 `Up` | `data/scraperbot.db` + `schema_migrations` | — |
| GORM Gen | `make gen` | `internal/model/*.gen.go`, `internal/query/` | 禁止 |
| Wails bindings | `make bindings` | `frontend/bindings/` | 禁止 |
| Wire | `make wire` | `internal/app/wire_gen.go` | 禁止 |
| DDL 参照 | 手動同期 | `storage/schema.sql` | 可（migrations と一致させる） |

## よく使うコマンド

| 操作 | コマンド |
|------|----------|
| 開発起動 | `make dev` |
| 本番ビルド | `make build` |
| テスト | `make test` |
| lint + test | `make check` |
| 依存整理 | `make go-tidy` |
| 開発 CLI 導入 | `make tools`（`go tool dlv` / `migrate` / `wails3`） |
| マイグレーション適用 | `make migrate-up` |
| DB バージョン確認 | `make migrate-version` |
| GORM Gen | `make gen` |
| Wire | `make wire` |
| TS bindings | `make bindings` |
