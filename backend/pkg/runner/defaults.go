package runner

// パッケージ既定の debug ログ付き実装。
var (
	defaultCrawler              Crawler              = NewCrawlerWithDebugLog(crawlerImpl{})
	defaultScraper              Scraper              = NewScraperWithDebugLog(scraperImpl{})
	defaultRobotsFetcher        RobotsFetcher        = NewRobotsFetcherWithDebugLog(robotsFetcherImpl{})
	defaultUIConfigurator       UIConfigurator       = NewUIConfiguratorWithDebugLog(uiConfiguratorImpl{})
	defaultFetchLimiterPreparer FetchLimiterPreparer = NewFetchLimiterPreparerWithDebugLog(fetchLimiterPreparerImpl{})
	defaultChromiumShutdown     ChromiumShutdown     = NewChromiumShutdownWithDebugLog(chromiumShutdownImpl{})
)
