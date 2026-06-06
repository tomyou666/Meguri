---
name: go-wire
description: Guides Google Wire dependency injection — composition root layout, Provide* providers, wireinject files, wire_gen workflow, and cleanup patterns. Use when editing Wire, adding dependencies, regenerating wire_gen, or working in the project's composition root package.
disable-model-invocation: true
---

# Go Wire

Wire の定義は **composition root パッケージ1か所** に閉じる（例: `internal/app`）。他パッケージに `wireinject` を増やさない。

## ファイル役割（例）

プロジェクトごとに名前は異なる。典型的な分割:

| ファイル（例） | 役割 |
|----------------|------|
| ルート struct 定義 | 実行時に束ねる依存（例: `Application`） |
| `providers.go` | `Provide*`、Wire 用アダプタ型 |
| `wire.go` | `//go:build wireinject` + injector + `wire.Build` リストのみ |
| `wire_gen.go` | 生成物。**手編集禁止** |

- `wire.go` に Provider 実装を書かない。
- composition root に **ドメイン・ビジネスロジックを置かない** — 下位パッケージ（例: インフラ層・ドメイン層・ユースケース層など、プロジェクトのレイアウトに従う）へ委譲する。

## 固定ルール

1. **composition root は1パッケージ** — Wire グラフの組み立てはそこだけ。
2. **Provider 名は `ProvideXxx`** — 下位の `NewXxx` と区別する。
3. **グラフの型は基本具象** — `wire.Bind` は interface 差し替えが必要なときだけ。
4. **組み立ては `Provide*` が原則** — 複数フィールドをそのまま注入するだけの型は `wire.Struct(new(T), "*")`。
5. **リソースの teardown** — ライフサイクルがある Provider は `(T, func(), error)` を返してよい。呼び出し側は **injector の第2戻り値 `cleanup` だけ** を `defer` する（Provider ごとの `func()` は Wire が連結する）。
6. **`wire_gen.go` はコミット** — `providers.go` / `wire.go` 変更後は Wire を再生成し、生成差分もコミットする（プロジェクトの Makefile・`go generate` 等に従う）。

## 都度判断（固定しない）

- **ルート struct のフィールド** — エントリポイント（CLI・HTTP・GUI 等）が何を必要とするかで追加。中間依存を載せるかはケースごと。
- **`Provide*` 内のロジック量** — 薄い `New*` ラッパーから、`Init` / `Close` など composition まで root に書く場合もある。ビジネス処理は下位パッケージへ。
- **`wire.Struct` vs 新規 `Provide*`** — フィールド注入だけなら `wire.Struct`、それ以外は `Provide*` を優先しつつケースで決める。

## Injector

- エントリ injector は **基本1本**（例: `Initialize`）。別エントリが必要なら injector を分けるか、引数で分岐するかは都度判断。
- 引数は **必要に応じて injector に追加**（例: `context.Context`、設定 struct）。
- 戻り値の典型: `( *RootStruct, func(), error )` — 第2戻り値が cleanup。

`wire.go` の injector 本体は `return nil, nil, nil`（または型に合うゼロ値）のスタブ。`wire.Build(...)` に列挙する。

```go
//go:build wireinject

//go:generate go run github.com/google/wire/cmd/wire

func Initialize(ctx context.Context, cfg *Config) (*Application, func(), error) {
    wire.Build(
        wire.Struct(new(Application), "*"),
        ProvideSomething,
        wire.Struct(new(SomeAdapter), "*"),
    )
    return nil, nil, nil
}
```

## 依存追加チェックリスト

1. 実装を composition root **以外** の適切なパッケージに置く。
2. `providers.go` に `ProvideXxx` を追加（必要ならアダプタ型も同ファイル）。
3. ルート struct に載せるか判断（エントリポイントから直接使うか）。
4. `wire.go` の `wire.Build` に `ProvideXxx`（または `wire.Struct`）を追加。
5. Wire を再生成 — エラーなら Provider の引数・戻り値型を修正。
6. `wire_gen.go` をコミットに含める。
7. プロジェクトに docstring 規約があればそれに従う。

## パターン例

**薄い Provider**

```go
func ProvideService(repo *Repository) *Service {
    return NewService(repo)
}
```

**ライフサイクル + cleanup**

```go
func ProvideClient(ctx context.Context, cfg *Config) (*Client, func(), error) {
    c := NewClient(cfg)
    if err := c.Connect(ctx); err != nil {
        return nil, nil, err
    }
    return c, func() { _ = c.Close(ctx) }, nil
}
```

**アダプタ型 + `wire.Struct`**

composition 専用の接着型は `providers.go` に置き、`wire.Struct(new(Adapter), "*")` で注入。

**Provider 戻り値が interface** — `wire.Bind` なしでも可。複数実装の差し替えが必要なときだけ `wire.Bind` を検討。

## 禁止・注意

- `wire_gen.go` を手で直さない。
- composition root **以外** に `wireinject` を置かない。
- cleanup をルート struct のフィールドに持たせず、injector の `cleanup` に集約する。
- `wire.go` と `wire_gen.go` のビルドタグ（`wireinject` / `!wireinject`）を崩さない。

## コード生成

プロジェクトの手順に従う（例: `make wire`、`go generate ./...`、`cd <composition-root> && go run github.com/google/wire/cmd/wire`）。README や Makefile を確認する。
