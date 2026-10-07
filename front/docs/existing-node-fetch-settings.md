# mode 3 / 4 の取得設定

mode 3（選択ノードから既存ノードのみ）と mode 4（選択ノードのみ取得）では、取得の並列・間隔・robots・同時取得上限が実行に乗らない。原因と、未実装の変更方針をまとめる。ソースはまだ変えていない。

## 結論（採用予定）

訪問リストの作り方は今のまま。取得だけを共有関数 `scrapeExistingNodesInOrder` のワーカープールにする。

ジョブ開始時の run 設定（app → workspace。ノード content は載せない）から、次を1回読む。

- `request_delay`
- `max_concurrency`
- `respect_robots_txt`
- `crawl.fetch_limits`（`http_max_inflight` / `chromium_max_inflight` / `auto_calibrate` / `dynamic_chromium`）

`request_delay` が 0 より大きいときは同時実行を 1 にする。待機は、実際にページを取得したあとのみ。除外・再取得スキップ・robots スキップでは待たない。

robots の Disallow は取得せずスキップする。robots.txt の取得失敗・パース失敗は許可して取得する。判定は新しいクロール API ではなく、runner に足す Allowed 判定から行う。ジョブの `FetchLimiter` と User-Agent を共有する。

各ノードの取得本体は今の `scrapeOneNode`（ノード content の上書き、`Crawl.Enabled=false`）。完了イベントは終わった順。プールはリストの先頭から、空いた枠の分だけ取り出す。

mode 3 のあとの manual pass は順次のまま。リンクは辿らない。探索系の設定は mode 3 / 4 では今どおり見ない。

## いま効かない理由

mode 1 / 2 は `runMainBFS` が `Crawl.Enabled=true` で `runner.CrawlWithProgress` を呼ぶ。ワーカー数・遅延・robots は `Crawler.Run` の中にある。

mode 3 / 4 は `scrapeExistingNodesInOrder` がノードを for で直列に回し、各ノードを `scrapeOneNode` → `runner.ScrapeWithConfig` で取り切ってから次へ進む。`scrapeOneNode` は `cfg.Crawl.Enabled = false` にして単発取得に落とす。この経路は `Crawler.Run` に入らない。

| 設定 | 読んでいる場所 | mode 3 / 4 |
| --- | --- | --- |
| `max_concurrency` | `Crawler.Run` のワーカー数。既定 4 | 単発取得が1件ずつ終わるまで次を始めない |
| `request_delay` | 同じワーカーループ。0 より大きいときワーカー数を 1 にし、取得のあと待つ | 待たない |
| `respect_robots_txt` | `Crawler.enqueue`。既定 true。Disallow は skip 理由 `robots` | 判定しない |
| `http_max_inflight` | `FetchLimiter`。既定 16。HTTP 取得の同時数 | ジョブ開始時に作るが、同時取得が 1 なので頭打ちにならない |
| `chromium_max_inflight` | 同上。既定 2 | 同上 |
| `auto_calibrate` | ジョブ開始時に Chromium 上限を測って上書き。既定オン | 開始時には走る。同時取得が 1 なので上限の差がスループットに出ない |
| `dynamic_chromium` | 実行中にメモリで Chromium 上限を ±1。既定オン | 同上 |

`runCrawl` はモード分岐の前に `limiterConfig()`（app → workspace）で `PrepareFetchLimiter` を1回呼び、`opts.FetchLimiter` と `opts.Cache` に載せる。mode 3 / 4 の `ScrapeWithConfig` も、キャッシュがあるときはその同じ limiter を Kernel に渡す。足りないのは同時に複数ノードを走らせる側である。

ノード JSON は実行前に content（`formats` 除く）だけ残る。`request_delay` や `fetch_limits` はノードでは上書きされない。読むなら run 設定を1回で足りる。ページ本文の selector などは、今どおりノードごとの `mergedConfig` が `scrapeOneNode` に残る。

UI の設定変換（`ParseUIConfig`）は `Config.Validate` を呼ばない。`Validate` 内の「`request_delay` が正なら `max_concurrency` を 1 にする」は、画面からの実行では走らない。mode 1 / 2 で同時実行が 1 になるのは、`Crawler.Run` が同じ条件を独自に見ているため。プール側でも同じ条件を見る。

## 訪問リスト（変えない）

mode 3 は選択ノードに、`ForwardReachableExisting` の結果（既存エッジで到達できるノード、BFS 順）を足す。HTML のリンク探索ではない。

mode 4 は指定 `nodeIds` のうち、グラフに存在する ID を入力順に並べる。

どちらのリストも、次を取得対象から外す処理は今のまま前段に置く。

- 再取得オフかつ status が `success`（理由 `duplicate_existing`）
- ワークスペース除外 URL、またはノードの `CrawlExclude`（status `skipped`、理由 `exclude_urls`）

キャンセル（`ctx` が閉じたとき）は今と同じく、その場でランを止める。1件の取得失敗では残りのノードを続ける。

mode 3 のあと、`runCrawl` はワークスペースの全ノードを到達済みにして `runManualPass` を呼ぶ。到達済みは manual pass 側で飛ばすので、この pass はほぼ動かない。今回も順次ループのまま触らない。

## 変更箇所

1. `front/internal/usecase/wails_service/scraper_service.go` の `scrapeExistingNodesInOrder`
   - 前段フィルタのあと、残りのノードをワーカープールで `scrapeOneNode` する。
   - ワーカー数は run 設定の `max_concurrency`（1〜64、未設定なら既定 4）。`request_delay > 0` なら 1。
   - 遅延の `time.After` は、スキップでない取得が終わったあと。ワーカーが複数いるときは遅延を掛けない（同時実行 1 のときだけ間隔になる）。
   - `enqueued` / `succeeded` / `failed` / `skipped` は複数 goroutine から触るので、加算を直列化する。
   - 取り出しはリスト順。完了通知は終わった順でよい。

2. runner の Allowed 判定（新規、クロールモードではない）
   - 入力は対象 URL、run 設定、ジョブの `RunOptions`（既存の `FetchLimiter`）。
   - 中身は BFS と同じ `robots.Cache`。User-Agent は `Plugins.Stealth.HTTP.EffectiveUserAgent()`。
   - robots.txt の GET はページ取得と同じ `FetchLimiter` を通す（HTTP 枠を消費する）。
   - ホスト単位キャッシュ。取得失敗・パース失敗は許可。
   - `respect_robots_txt` が false のときは呼ばない。
   - Disallow のノードは `scrapeOneNode` に入らない。`skipped` を加算し、status `skipped`、理由 `robots`。遅延は掛けない。

3. `scrapeOneNode` は変えない。`Crawl.Enabled=false` のままなので、返ってきたリンクから新規ノードは作らない。

`FetchLimiter` の生成・`auto_calibrate`・`dynamic_chromium` の開始は `runCrawl` の既存処理に任せる。ノードごとに作り直さない。

## 変えない設定

次は mode 3 / 4 の訪問リストに掛けない。`Crawler.Run` にシードとして流すと、これらが突然効く。

- `max_depth`
- `max_pages`
- `allow_external_links`
- `allow_subdomains`
- `include_paths` / `exclude_paths`
- `include_hosts` / `exclude_hosts`

`exclude_urls` とノードの `CrawlExclude` は、今の前段スキップのまま（探索フィルタとして Crawler に渡すのではない）。

## 実装時の注意

- 同一設定ハッシュの Kernel は `ScrapeCache` で共有される。mode 1 / 2 のワーカーが既に同じパイプラインを並列で使っている。mode 3 / 4 もそのキャッシュ経由に乗せる。
- 永続化は SQLite（WAL、`busy_timeout` 5000ms）。`database/sql` の接続は goroutine から呼んでよい。書き込みが込み合ったときの待ちは、この timeout の範囲に収まる想定で確認する。
- mode 4 のコメントは「入力順に scrape」となっている。対象集合の並びは入力順のまま、完了は終わった順になる、とコメントを直す。
- 画面用の `FetchRobotsTxt` は本文を返すだけで、Disallow 判定・失敗時の許可・ホストキャッシュ・ジョブの limiter 共有がない。Allowed 判定の代わりに使わない。
