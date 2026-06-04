---
name: go-wire
description: Guides Google Wire dependency injection in scraper-bot backend internal/app — composition root layout, Provide* providers, wire.go injectors, wire_gen workflow, and cleanup patterns. Use when editing Wire, adding dependencies, running make wire, or working in backend/internal/app.
disable-model-invocation: true
---

# Go Wire (scraper-bot backend)

`backend/internal/app` が唯一の composition root。Wire の変更はこのパッケージに閉じる。

## ファイル役割

| ファイル | 役割 |
|----------|------|
| `app.go` | `Application` 構造体 |
| `providers.go` | `Provide*`、Wire 用アダプタ型（例: `FileResultSink`） |
| `wire.go` | `//go:build wireinject` + `Initialize` + `wire.Build` リストのみ |
| `wire_gen.go` | 生成物。**手編集禁止** |

`wire.go` に Provider 実装を書かない。`providers.go` にビジネスロジックを置かない（`core` / `infrastructure` / `usecase` へ）。

## 固定ルール

1. **composition root は `internal/app` のみ** — 他パッケージに `wireinject` を増やさない。
2. **Provider 名は `ProvideXxx`** — 下位の `NewXxx` と区別する。
3. **グラフの型は基本具象** — `wire.Bind` は interface 差し替えが必要なときだけ。
4. **組み立ては `Provide*` が原則** — 複数フィールドをそのまま注入するだけの型は `wire.Struct(new(T), "*")`（例: `Application`, `FileResultSink`）。
5. **リソースの teardown** — ライフサイクルがある Provider は `(T, func(), error)` を返してよい。呼び出し側は **`Initialize` の第2戻り値 `cleanup` だけ** を `defer` する（Provider ごとの `func()` は Wire が連結する）。
6. **`wire_gen.go` はコミット** — `providers.go` / `wire.go` 変更後は `backend` で `make wire` し、生成差分もコミットする。

## 都度判断（固定しない）

- **`Application` のフィールド** — presentation が何を必要とするかで追加する。中間依存を載せるかはケースごと。
- **`Provide*` 内のロジック量** — 薄い `New*` ラッパーから、`Init` / `Close` など composition まで app に書く場合もある（`ProvideKernel` 参照）。ドメイン・ビジネス処理は下位層へ。
- **`wire.Struct` vs 新規 `Provide*`** — 上記「原則」と現状の整合で決める。

## Injector

- エントリは **`Initialize` 1本**（別 GUI 用 injector は増やさない方針）。
- 引数は **必要に応じて `Initialize` に追加**（現状: `ctx context.Context`, `cfg *model.Config`）。
- 戻り値: `(*Application, func(), error)`。

`wire.go` の `Initialize` 本体は `return nil, nil, nil` のスタブ。`wire.Build(...)` に列挙する。

```go
//go:build wireinject

//go:generate go run github.com/google/wire/cmd/wire

func Initialize(ctx context.Context, cfg *model.Config) (*Application, func(), error) {
    wire.Build(
        wire.Struct(new(Application), "*"),
        ProvideSomething,
        wire.Struct(new(SomeAdapter), "*"),
    )
    return nil, nil, nil
}
```

## 依存追加チェックリスト

1. 実装を `infrastructure` / `core` / `usecase` に置く（app にロジックを増やさない）。
2. `providers.go` に `ProvideXxx` を追加（必要ならアダプタ型も同ファイル）。
3. `app.go` の `Application` に載せるか判断（presentation から使うか）。
4. `wire.go` の `wire.Build` に `ProvideXxx`（または `wire.Struct`）を追加。
5. `backend` で `make wire` — エラーなら Provider の引数・戻り値型を修正。
6. `wire_gen.go` をコミットに含める。
7. docstring は [go-docstring-style](../go-docstring-style/SKILL.md) に従う。

## よくあるパターン（現状コード）

**薄い Provider**

```go
func ProvideScrape(pipeline *core.Pipeline) *usecase.Scrape {
    return usecase.NewScrape(pipeline)
}
```

**ライフサイクル + cleanup**

```go
func ProvideKernel(...) (*core.Kernel, func(), error) {
    k := core.NewKernel(...)
    if err := k.Init(ctx); err != nil {
        return nil, nil, err
    }
    return k, func() { _ = k.Close(ctx) }, nil
}
```

**アダプタ型 + `wire.Struct`**

`FileResultSink` のように app 専用の接着型は `providers.go` に置き、`wire.Struct(new(FileResultSink), "*")` で注入。

**Provider 戻り値が interface** — `wire.Bind` なしでも可（例: `ProvideHost` → `plugin.Host`）。interface への束ねが複数実装あるときだけ `wire.Bind` を検討。

## 禁止・注意

- `wire_gen.go` を手で直さない。
- `presentation` / `usecase` / `infrastructure` に `wireinject` を置かない。
- cleanup を `Application` フィールドに持たせず、`Initialize` の `cleanup` に集約する。
- `wire.go` と `wire_gen.go` のビルドタグ（`wireinject` / `!wireinject`）を崩さない。

## コマンド

```bash
cd backend && make wire
```

README: `backend/README.md`（Wire 節）。
