export const messages = {
	appName: 'Meguri',
	version: '0.1.0',
	toast: {
		dismissAria: '通知を閉じる',
	},
	update: {
		upToDate: '最新バージョンです',
		updateReady: (version: string) =>
			`バージョン ${version} の更新をダウンロードしました。アプリを再起動して適用してください。`,
		updateReadyNoVersion:
			'更新をダウンロードしました。アプリを再起動して適用してください。',
		unavailable: '更新機能は利用できません',
		checkFailed: '更新の確認に失敗しました',
	},
	menu: {
		file: 'ファイル',
		openScrb: '開く (.crawlproj)',
		saveScrb: '保存 (.crawlproj)',
		openFileMenu: 'ファイルメニューを開く',
		settings: '設定',
		appDefaults: 'デフォルト設定',
		openSettingsMenu: '設定メニューを開く',
		feedback: 'フィードバック',
		openFeedback: 'フィードバックを送る',
		checkForUpdates: '更新を確認…',
		mergeAll: '全結果をマージ',
		mergeSelected: '選択マージ',
		exportAll: '全結果エクスポート',
		exportSelected: '選択エクスポート',
	},
	control: {
		play: '再生',
		playMode1: '再生',
		playMode2: '選択ノードから再生',
		playMode3: '既存ノードのみ再生',
		playMode4: '選択ノードのみ再生',
		pause: '一時停止',
		stop: '停止',
		formats: '保存形式',
		mode1: '起点 URL から開始',
		mode2: '選択ノードから（デフォルト設定）',
		mode3: '選択ノードから既存ノードのみ',
		mode4: '選択ノードのみ取得（リンク探索なし）',
		closeModeMenu: '実行モードメニューを閉じる',
		rescrapeExisting: '既存ノードを再取得',
	},
	sidebar: {
		workspaces: 'ワークスペース',
		domainStatus: 'ドメインステータス',
		newWorkspace: '新規ワークスペース',
		newNode: 'ノード追加',
		deleteNode: 'ノード削除',
		emptyDomains: 'ドメインがありません',
		emptyWorkspaces: 'ワークスペースがありません',
		openLeft: '左サイドバーを開く',
		closeLeft: '左サイドバーを閉じる',
		openRight: '右サイドバーを開く',
		closeRight: '右サイドバーを閉じる',
		workspaceSettings: 'WS 設定',
		diffSummary: '差分サマリ',
		deleteWorkspace: 'ワークスペースを削除',
		duplicateWorkspace: 'ワークスペースをコピー',
		openWorkspaceMenu: 'ワークスペースのその他操作',
		diffCompare: '差分を表示',
		diffCountAria: (count: number) => `差分 ${count} 件`,
	},
	diff: {
		summaryTitle: '差分サマリ',
		filterAll: 'すべて',
		kindContent: 'content',
		kindLinks: 'links',
		kindFetch: 'fetch',
		emptyNodes: '差分ノードはありません',
		loading: '読み込み中…',
		toastTitle: (count: number) => `差分を検出しました（${count} ノード）`,
		toastAction: '詳細を見る',
		baselineUpdated: '変更を確認済みにしました',
		baselineUpdateFailed: '変更の確認に失敗しました',
		markReviewed: '変更を確認済みにする',
		markReviewedAria: (count: number) =>
			`未確認の差分 ${count} 件を確認済みにする`,
		windowTitle: 'ノード差分',
		openWindowFailed: '差分ウィンドウを開けませんでした',
		fetchState: {
			success: '成功',
			error: '失敗',
			skipped: 'スキップ',
			none: '未取得',
		},
		tooltipContent: '本文の差分',
		tooltipLinks: 'リンクの差分',
		tooltipFetch: '取得状態の差分',
	},
	domainStatus: {
		statusLabel: (status: string, count: number) => {
			const labels: Record<string, string> = {
				success: '成功',
				error: '失敗',
				skipped: 'スキップ',
				running: '実行中',
				idle: '待機',
			};
			return `${labels[status] ?? status} ${count}`;
		},
		statusSummary: (counts: {
			success: number;
			error: number;
			skipped: number;
			running: number;
			idle: number;
		}) => {
			const parts: string[] = [];
			if (counts.success > 0) parts.push(`成功 ${counts.success}`);
			if (counts.error > 0) parts.push(`失敗 ${counts.error}`);
			if (counts.skipped > 0) parts.push(`スキップ ${counts.skipped}`);
			if (counts.running > 0) parts.push(`実行中 ${counts.running}`);
			if (counts.idle > 0) parts.push(`待機 ${counts.idle}`);
			return parts.join(' / ') || '—';
		},
		robotsLoading: 'robots 取得中…',
		robotsFound: 'robots あり',
		robotsNotFound: 'robots なし',
		robotsError: '取得失敗',
		robotsEmpty: '（空）',
		robotsNotFoundDetail: (code?: number) =>
			code
				? `robots.txt は見つかりませんでした（HTTP ${code}）`
				: 'robots.txt は見つかりませんでした',
		robotsErrorDetail: (error?: string, code?: number) => {
			const parts: string[] = [];
			if (code) parts.push(`HTTP ${code}`);
			if (error) parts.push(error);
			return parts.length > 0
				? `robots.txt の取得に失敗しました（${parts.join(' / ')}）`
				: 'robots.txt の取得に失敗しました';
		},
	},
	right: {
		runSummary: '実行サマリ',
		nodeResult: 'ノード結果',
		nodeSettings: 'ノード設定',
		noSelection: 'ノードを選択するか、クロールを実行してください',
		crawlExclude: 'このノードと配下をクロールしない',
		history: '実行履歴',
		multiSelectCount: (n: number) => `${n} ノード選択中`,
		preview: 'プレビュー',
		save: '保存',
		delete: '削除',
		runModeBadge: (mode: number) => `モード ${mode}`,
		runStats: (succeeded: number, failed: number, skipped: number) =>
			`成功 ${succeeded} / 失敗 ${failed} / スキップ ${skipped}`,
		runStatsDuplicateLinks: (count: number) => `重複リンクスキップ ${count}`,
		crawlLog: 'クロールログ',
		crawlLogEmpty: 'ログはまだありません',
		linkSkipReason: (reason: string) =>
			reason === 'duplicate_existing'
				? '既存URL'
				: reason === 'duplicate_in_run'
					? '実行内重複'
					: reason,
		linkSkipLine: (parent: string, target: string, reason: string) =>
			parent ? `${parent} → ${target}（${reason}）` : `${target}（${reason}）`,
		noResultApi: '結果がありません（API未取得）',
		transformerBadge: (name: string) => `transformer: ${name}`,
		manuallyEdited: '手動編集',
		copy: 'コピー',
		copied: 'クリップボードにコピーしました',
		copyFailed: 'コピーに失敗しました',
		maximize: '拡大表示',
		maximizeFailed: '拡大表示ウィンドウを開けませんでした',
		edit: '編集',
		source: 'ソース',
		previewLabel: 'プレビュー',
		updateSaved: '結果を保存しました',
		updateFailed: '結果の保存に失敗しました',
	},
	dialog: {
		newWorkspaceTitle: '新規ワークスペース',
		newWorkspaceName: '名前',
		newWorkspaceUrl: '起点 URL',
		addNodeTitle: 'ノード追加',
		addNodeUrl: 'URL',
		deleteNodeTitle: 'ノード削除',
		deleteNodeConfirm: '選択ノードと配下のノードを削除しますか？',
		deleteWorkspaceTitle: 'ワークスペース削除',
		deleteWorkspaceConfirm:
			'このワークスペースを削除しますか？関連するグラフ・結果もすべて削除されます。',
		duplicateWorkspaceTitle: 'ワークスペースコピー',
		duplicateWorkspaceName: '名前',
		cancel: 'キャンセル',
		create: '作成',
		copy: 'コピー',
		add: '追加',
		delete: '削除',
		confirm: '確認',
	},
	settings: {
		save: '保存',
		saving: '保存中…',
		saveSuccess: '設定を保存しました',
		saveFailed: '設定の保存に失敗しました',
		validationFailed: '入力内容に誤りがあります',
		tabs: {
			general: '基本',
			request: 'HTTP',
			content: '本文',
			pdf: 'PDF',
			crawl: 'クロール',
			plugins: '取得方法',
			output: '出力先',
		},
		help: {
			timeout:
				'1ページ取得の最大待ち時間（例: 30s）。長すぎると全体が遅くなります。',
			retry_count: '失敗時に再試行する回数（0〜10）。',
			retry_interval: '再試行までの待ち時間（例: 2s）。',
			userAgent: 'サーバーに送るブラウザ名。ブロック対策や識別に使います。',
			formats:
				'保存する形式。markdown / links など複数選択できます（1つ以上必須）。',
			only_main_content: '広告やナビを除き、記事本文だけを抽出します。',
			include_tags: '抽出対象の HTML タグ（1行1つ）。空なら自動判定。',
			exclude_tags: '除外する HTML タグ（1行1つ）。',
			selector: '本文を取る CSS セレクタ。空ならページ全体から推定。',
			extract_links: 'ページ内のリンク一覧を別形式で保存します。',
			extract_metadata: 'タイトル・description などのメタ情報を保存します。',
			pdf_enabled: 'PDF URL をテキスト化して取得するかどうか。',
			pdf_mode: 'fast=速い / auto=自動 / ocr=スキャン向け。',
			pdf_max_pages: '1ファイルあたり読む最大ページ数。',
			pdf_output: 'PDF から得るテキストの形式。',
			crawl_enabled: 'この設定層でリンクを辿って巡回するか。',
			max_depth: '起点から何階層まで辿るか（0=起点のみ）。',
			max_pages: 'ワークスペース全体で訪問する最大ページ数。',
			include_paths: '辿る URL パス（1行1つ。例: /docs）。空=制限なし。',
			exclude_paths: '辿らない URL パス（1行1つ）。',
			allow_external_links: '別ドメインへのリンクも辿るか。',
			allow_subdomains: 'サブドメイン（例: blog.example.com）も辿るか。',
			request_delay: 'リクエスト間隔（例: 500ms）。サーバー負荷軽減用。',
			max_concurrency:
				'同時に走らせるクロールワーカー数（1〜64）。実際の取得数は fetch_limits で別途制限される。',
			fetch_limits_overview:
				'max_concurrency はワーカー数、下の項目は HTTP/Chromium の実際の同時取得数。Chromium は auto_calibrate → dynamic_chromium の順で上限が変わる。',
			http_max_inflight:
				'HTTP の同時取得上限（1〜64、既定 16）。auto_calibrate / dynamic_chromium の対象外で常にこの値。',
			chromium_max_inflight:
				'Chromium の静的上限（1〜8、既定 2）。両方オフなら常にこの値。オンなら起動時の再計算の基準・失敗時のフォールバック・動的調整の戻り上限になる。',
			fetch_auto_calibrate:
				'オン: ジョブ開始時にブラウザを1回起動してメモリ計測し、上限を1〜8で上書き（失敗時は2）。オフ: 左の chromium_max_inflight をそのまま使う。',
			fetch_dynamic_chromium:
				'オン: 5秒ごとにメモリ使用率を見て上限を±1（高水位0.8超で減、低水位0.6未満で増、最小1）。オフ: 起動時の上限をジョブ中固定。',
			respect_robots_txt: 'robots.txt の Disallow を守るか。',
			fetcher: 'http=軽量 / chromium=JavaScript 必須ページ向け（重い）。',
			transformer:
				'本文の変換方式。markdown / html / raw_html / json から選択。プレビューとファイル出力の主形式になります。',
			browser_path: 'Chromium 実行ファイルのパス。空なら自動検出。',
			headless: '画面を表示せずバックグラウンドでブラウザを動かす。',
			output_dir: '結果ファイルを書き出すフォルダ。',
			file_pattern: 'ファイル名のテンプレート（例: {url_hash}.md）。',
			ws_name: '一覧に表示するワークスペース名。',
			seed_url: 'グラフの起点となる URL。新規 WS 作成時に使います。',
		},
	},
	error: {
		globalBanner: 'アプリケーションエラー',
		crawlFailed: 'クロールエラー',
		nodeFailed: 'ノードエラー',
		deleteWorkspaceFailed: 'ワークスペースの削除に失敗しました',
		duplicateWorkspaceFailed: 'ワークスペースのコピーに失敗しました',
	},
	graph: {
		layout: 'レイアウト',
		layoutVertical: '縦方向に自動配置',
		layoutHorizontal: '横方向に自動配置',
		expandAll: 'すべて展開',
		collapseAll: 'すべて折りたたむ',
		expandDetail: 'ノード詳細を展開',
		collapseDetail: 'ノード詳細を折りたたむ',
		expandSubtree: '配下を展開',
		collapseSubtree: '配下を折りたたむ',
		zoomIn: 'ズームイン',
		zoomOut: 'ズームアウト',
		fitView: '全体表示',
		contextCollapse: '折りたたむ',
		contextExpand: '展開',
		contextExcludeCrawl: 'クロールしない',
		contextScrape: 'スクレイプ',
		contextPreviewResult: '結果プレビュー',
		contextDelete: '削除',
		toolPan: '手のひら（パン）— ドラッグで移動',
		toolSelect: '矩形選択 — ドラッグで範囲選択（Ctrl で追加）',
		minimapOpen: 'ミニマップを表示',
		minimapTitle: 'ミニマップ',
		minimapClose: 'ミニマップを閉じる',
	},
	status: {
		idle: '未訪問',
		running: '実行中',
		success: '成功',
		error: '失敗',
		skipped: 'スキップ',
	},
	bootstrapLoading: '読み込み中…',
	export: {
		windowTitleAll: '全結果エクスポート',
		windowTitleSelected: '選択エクスポート',
		orderTitle: 'マージ順',
		settingsTitle: 'エクスポート設定',
		previewTitle: 'プレビュー',
		previewEmpty: '「プレビュー開始」を押すとここに表示されます',
		previewStart: 'プレビュー開始',
		previewLoading: 'プレビュー取得中…',
		save: 'ファイルに保存',
		copy: 'クリップボードにコピー',
		selectAll: '全選択',
		deselectAll: '全解除',
		cascadeCheck: '連動選択',
		splitSave: '分割保存（ZIP でまとめる）',
		splitSaveHint: 'チェック済みノードごとにファイルを作成し ZIP で保存します',
		format: '出力形式',
		formatMarkdown: 'Markdown',
		formatHtml: 'HTML',
		separator: '区切り文字',
		separatorHint: '\\n \\r\\n \\t などのエスケープが使えます',
		includeHeading: '見出しを付与',
		headingField: '見出しに使う項目',
		headingUrl: 'URL',
		headingLabel: 'ラベル',
		noNodesChecked: 'エクスポート対象のノードを選択してください',
		noNodesInTree: 'エクスポート対象のノードがありません',
		skippedNoResult: (n: number) => `結果がないノード ${n} 件を除外しました`,
		copied: 'クリップボードにコピーしました',
		copyFailed: 'コピーに失敗しました',
		saveFailed: 'ファイルの保存に失敗しました',
		saveSuccess: 'ファイルを保存しました',
		saveZipSuccess: 'ZIP ファイルを保存しました',
		openFailed: 'エクスポートウィンドウを開けませんでした',
	},
} as const;
